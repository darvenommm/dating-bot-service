package bot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/telebot.v3"

	actionv1 "github.com/darvenommm/dating-bot-service/pkg/api/action/v1"
	commonv1 "github.com/darvenommm/dating-bot-service/pkg/api/common/v1"
	filterv1 "github.com/darvenommm/dating-bot-service/pkg/api/filter/v1"
	matchv1 "github.com/darvenommm/dating-bot-service/pkg/api/match/v1"
	profilev1 "github.com/darvenommm/dating-bot-service/pkg/api/profile/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	cmdStart     = "/start"
	cmdFilter    = "/filter"
	cmdRecommend = "/recommend"
	cmdCancel    = "/cancel"
	cmdHelp      = "/help"
	cmdSkip      = "/skip"

	maxDescriptionLength = 500
	maxPhotoSize         = 5 * 1024 * 1024
)

var (
	mu            sync.Mutex
	profileStates = make(map[int64]*ProfileState)
	filterStates  = make(map[int64]*FilterState)
	recStates     = make(map[int64]*profilev1.Profile)
	filterDone    = make(map[int64]bool)

	likeBtn    = telebot.InlineButton{Unique: "like", Text: "Like"}
	dislikeBtn = telebot.InlineButton{Unique: "dislike", Text: "Dislike"}
)

type ProfileState struct {
	step        int
	fullName    string
	gender      commonv1.Gender
	age         uint32
	description string
	photo       []byte
}

type FilterState struct {
	step   int
	gender commonv1.Gender
	minAge uint32
	maxAge uint32
}

func StartListeningBot(token string) error {
	conn, err := grpc.NewClient("127.0.0.1:10000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("gRPC dial error: %v", err)
	}
	defer conn.Close()

	profileClient := profilev1.NewProfileServiceClient(conn)
	filterClient := filterv1.NewFilterServiceClient(conn)
	actionClient := actionv1.NewActionServiceClient(conn)
	matchClient := matchv1.NewMatchServiceClient(conn)

	// Telegram bot
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return err
	}

	// Handlers
	bot.Handle(cmdHelp, func(c telebot.Context) error {
		return c.Send(
			"Доступные команды:\n" +
				"/start – создать профиль\n" +
				"/filter – настроить фильтр\n" +
				"/recommend – новые рекомендации\n" +
				"/cancel – отменить процесс\n" +
				"При вводе текста можно отправить /skip, чтобы пропустить необязательный шаг.",
		)
	})

	bot.Handle(cmdStart, func(c telebot.Context) error {
		log.Printf("User %d started creating profile", c.Sender().ID)
		uid := c.Sender().ID
		mu.Lock()
		profileStates[uid] = &ProfileState{step: 1}
		mu.Unlock()
		return c.Send("Введите ваше полное имя:")
	})

	bot.Handle(cmdFilter, func(c telebot.Context) error {
		uid := c.Sender().ID
		log.Printf("User %d is configuring filter", uid)

		// Проверка, создан ли профиль
		profileResp, err := profileClient.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: uid})
		if err != nil || profileResp == nil || profileResp.Profile == nil {
			log.Printf("Profile not found for user %d", uid)
			return c.Send("Сначала создайте профиль через /start.")
		}

		mu.Lock()
		filterStates[uid] = &FilterState{step: 1}
		mu.Unlock()
		return sendGenderKeyboard(c, "Выберите пол для поиска:")
	})

	bot.Handle(cmdRecommend, func(c telebot.Context) error {
		uid := c.Sender().ID
		log.Printf("User %d is requesting recommendations", uid)

		// Запрашиваем профиль пользователя
		profileResp, err := profileClient.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: uid})
		if err != nil || profileResp == nil || profileResp.Profile == nil {
			log.Printf("Error fetching profile for User ID: %d, Error: %v", uid, err)
			return c.Send("Сначала создайте профиль через /start.")
		}

		// Проверка, установлен ли фильтр
		mu.Lock()
		ok := filterDone[uid]
		mu.Unlock()
		if !ok {
			log.Printf("Filter is not set for User ID: %d", uid)
			return c.Send("Сначала настройте фильтр через /filter.")
		}

		// Запрашиваем рекомендации
		resp, err := profileClient.GetRecommendation(context.Background(), &profilev1.GetRecommendationRequest{UserId: uid})
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				// Если ошибка типа "NotFound", сообщаем, что нет рекомендаций
				log.Printf("No recommendations for User ID: %d", uid)
				return c.Send("Для вас пока нет рекомендаций. Попробуйте позже.")
			}
			log.Printf("GetRecommendation error: %v", err)
			return c.Send("Ошибка при получении рекомендаций.")
		}

		// Проверка на пустой ответ
		if resp == nil || resp.Profile == nil {
			log.Printf("Recommendation response is nil for User ID: %d", uid)
			return c.Send("Для вас пока нет рекомендаций. Попробуйте позже.")
		}

		// Сохраняем профиль для дальнейшей работы
		mu.Lock()
		recStates[uid] = resp.Profile
		mu.Unlock()

		log.Printf("Sending recommendation for User ID: %d", uid)

		// Создаем кнопки для лайка/дизлайка
		markup := &telebot.ReplyMarkup{}
		markup.Inline(markup.Row(
			telebot.Btn{Unique: "like", Text: "Like"},
			telebot.Btn{Unique: "dislike", Text: "Dislike"},
		))

		// Проверяем, есть ли фото в профиле
		if resp.Profile.Photo != nil && len(resp.Profile.Photo.Value) > 0 {
			// Если есть фото, отправляем фото с подписью
			photo := &telebot.Photo{
				File:    telebot.File{FileReader: bytes.NewReader(resp.Profile.Photo.Value)},
				Caption: fmt.Sprintf("%s, %d лет\n%s", resp.Profile.FullName, resp.Profile.Age, resp.Profile.Description.GetValue()),
			}
			return c.Send(photo, markup)
		}

		// Если фото нет, отправляем текст
		text := fmt.Sprintf("%s, %d лет\n%s", resp.Profile.FullName, resp.Profile.Age, resp.Profile.Description.GetValue())
		return c.Send(text, markup)
	})

	bot.Handle(cmdCancel, func(c telebot.Context) error {
		uid := c.Sender().ID
		log.Printf("User %d canceled the process", uid)
		mu.Lock()
		delete(profileStates, uid)
		delete(filterStates, uid)
		delete(recStates, uid)
		filterDone[uid] = false
		mu.Unlock()
		return c.Send("Все процессы сброшены. Начните снова с /start.")
	})

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		uid := c.Sender().ID
		text := c.Text()

		// если ждем оценки — просим нажать кнопку
		mu.Lock()
		_, waitingRec := recStates[uid]
		mu.Unlock()
		if waitingRec {
			return c.Send("Пожалуйста, воспользуйтесь кнопками для оценки.")
		}

		// шаги создания профиля
		mu.Lock()
		ps, inProfile := profileStates[uid]
		fs, inFilter := filterStates[uid]
		mu.Unlock()

		if inProfile {
			return handleProfileStep(c, profileClient, ps, text)
		}
		if inFilter {
			return handleFilterStep(c, filterClient, fs, text)
		}
		return nil
	})

	bot.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		uid := c.Sender().ID
		mu.Lock()
		ps, ok := profileStates[uid]
		mu.Unlock()
		if !ok || ps.step != 5 {
			// Если профиль не существует или не на нужном шаге, пропускаем обработку
			return nil
		}

		// Получаем фотографии из сообщения
		photo := c.Message().Photo
		if photo.FileSize == 0 || photo.FileSize > maxPhotoSize {
			return c.Send("Фото слишком большое. Максимальный размер — 5 МБ.")
		}

		file := photo.File

		// Получаем файл
		reader, err := c.Bot().File(&file)
		if err != nil {
			return c.Send("Не удалось загрузить фото, попробуйте ещё раз.")
		}
		defer reader.Close()

		// Читаем данные файла в буфер
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, reader)
		if err != nil {
			return c.Send("Не удалось загрузить фото.")
		}

		// Сохраняем фото в поле ProfileState
		ps.photo = buf.Bytes()

		// Завершаем профиль
		return finalizeProfile(c, profileClient, ps)
	})

	// колбэки лайк/дизлайк
	bot.Handle(&likeBtn, makeActionHandler(actionClient, actionv1.Action_ACTION_LIKE))
	bot.Handle(&dislikeBtn, makeActionHandler(actionClient, actionv1.Action_ACTION_DISLIKE))

	// слушаем матчи
	go listenMatches(bot, profileClient, matchClient)

	log.Println("Bot started")

	bot.Start()

	return nil
}

// ---- helpers ----

func sendGenderKeyboard(c telebot.Context, prompt string) error {
	m := &telebot.ReplyMarkup{ResizeKeyboard: true}
	btns := []telebot.Btn{
		m.Text("Мужской"),
		m.Text("Женский"),
	}
	m.Reply(m.Row(btns...))
	return c.Send(prompt, m)
}

func handleProfileStep(c telebot.Context, client profilev1.ProfileServiceClient, ps *ProfileState, text string) error {
	switch ps.step {
	case 1:
		// Валидация для полного имени
		if len(strings.Fields(text)) != 2 {
			return c.Send("Полное имя должно состоять из двух слов.")
		}
		ps.fullName = text
		ps.step = 2
		return sendGenderKeyboard(c, "Выберите ваш пол:")
	case 2:
		switch text {
		case "Мужской":
			ps.gender = commonv1.Gender_GENDER_MALE
		case "Женский":
			ps.gender = commonv1.Gender_GENDER_FEMALE
		}
		ps.step = 3
		return c.Send("Введите ваш возраст (числом):", &telebot.ReplyMarkup{RemoveKeyboard: true})
	case 3:
		age, err := strconv.Atoi(text)
		if err != nil || age < 18 || age > 100 {
			return c.Send("Возраст должен быть от 18 до 100.")
		}
		ps.age = uint32(age)
		ps.step = 4
		return c.Send("Расскажите о себе или /skip:")
	case 4:
		if text != cmdSkip && len(text) > maxDescriptionLength {
			return c.Send(fmt.Sprintf("Описание слишком длинное. Максимальная длина — %d символов.", maxDescriptionLength))
		}
		if text != cmdSkip {
			ps.description = text
		}
		ps.step = 5
		return c.Send("Пришлите фото или /skip:")
	case 5:
		if text == cmdSkip {
			return finalizeProfile(c, client, ps)
		}
		return c.Send("Ожидаю фото или /skip:")
	}
	return nil
}

func handleFilterStep(c telebot.Context, client filterv1.FilterServiceClient, fs *FilterState, text string) error {
	switch fs.step {
	case 1:
		// Устанавливаем пол для фильтра
		switch text {
		case "Мужской":
			fs.gender = commonv1.Gender_GENDER_MALE
		case "Женский":
			fs.gender = commonv1.Gender_GENDER_FEMALE
		}
		fs.step = 2
		return c.Send("Введите минимальный возраст:", &telebot.ReplyMarkup{RemoveKeyboard: true})
	case 2:
		// Устанавливаем минимальный возраст для фильтра
		minAge, err := strconv.Atoi(text)
		if err != nil {
			return c.Send("Пожалуйста, укажите возраст числом.")
		}
		if minAge < 18 || minAge > 100 {
			return c.Send("Минимальный возраст должен быть от 18 до 100 лет.")
		}
		fs.minAge = uint32(minAge)
		fs.step = 3
		return c.Send("Введите максимальный возраст:")
	case 3:
		// Устанавливаем максимальный возраст для фильтра
		maxAge, err := strconv.Atoi(text)
		if err != nil {
			return c.Send("Пожалуйста, укажите возраст числом.")
		}
		if uint32(maxAge) < fs.minAge {
			return c.Send("Максимальный возраст должен быть больше минимального возраста.")
		}
		if maxAge > 100 {
			return c.Send("Максимальный возраст не должен превышать 100 лет.")
		}
		fs.maxAge = uint32(maxAge)

		// Завершаем настройку фильтра, вызывая finalizeFilter
		userID := c.Sender().ID
		req := &filterv1.SetFilterRequest{
			UserId: userID,
			Gender: fs.gender,
			MinAge: fs.minAge,
			MaxAge: fs.maxAge,
		}

		// Отправляем запрос на установку фильтра
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := client.SetFilter(ctx, req); err != nil {
			log.Printf("SetFilter error: %v", err)
			return c.Send("Не удалось сохранить фильтр.")
		}

		// Завершаем процесс настройки фильтра
		mu.Lock()
		delete(filterStates, userID)
		filterDone[userID] = true
		mu.Unlock()

		// Вызов finalizeFilter после успешного сохранения фильтра
		return finalizeFilter(c, client, fs)
	}
	return nil
}

func finalizeProfile(c telebot.Context, client profilev1.ProfileServiceClient, ps *ProfileState) error {
	uid := c.Sender().ID
	req := &profilev1.SetProfileRequest{
		UserId:   uid,
		FullName: ps.fullName,
		Gender:   ps.gender,
		Age:      ps.age,
	}
	if ps.description != "" {
		req.Description = &ps.description
	}
	if len(ps.photo) > 0 {
		req.Photo = ps.photo
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.SetProfile(ctx, req); err != nil {
		log.Printf("SetProfile error: %v", err)
		return c.Send("Не удалось сохранить профиль.")
	}
	mu.Lock()
	delete(profileStates, uid)
	filterDone[uid] = false
	mu.Unlock()
	return c.Send("Профиль создан! Теперь настройте фильтр через /filter.")
}

func finalizeFilter(c telebot.Context, client filterv1.FilterServiceClient, fs *FilterState) error {
	userID := c.Sender().ID
	req := &filterv1.SetFilterRequest{
		UserId: userID, Gender: fs.gender,
		MinAge: fs.minAge, MaxAge: fs.maxAge,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.SetFilter(ctx, req); err != nil {
		log.Printf("SetFilter error: %v", err)
		return c.Send("Не удалось сохранить фильтр.")
	}
	mu.Lock()
	delete(filterStates, userID)
	filterDone[userID] = true
	mu.Unlock()
	return c.Send("Фильтр установлен! Для рекомендаций используйте /recommend.")
}

func makeActionHandler(client actionv1.ActionServiceClient, act actionv1.Action) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		log.Println(recStates)
		uid := c.Sender().ID
		mu.Lock()
		profile, ok := recStates[uid]
		mu.Unlock()
		if !ok {
			log.Printf("No active recommendation for User ID: %d", uid)
			return c.Send("Нет активных рекомендаций.") // Используем Send, а не Respond
		}
		req := &actionv1.AddActionRequest{
			FromUserId: uid,
			ToUserId:   profile.UserId,
			Action:     act,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := client.AddAction(ctx, req); err != nil {
			log.Printf("AddAction error: %v", err)
			return c.Send("Ошибка при отправке действия.") // Используем Send
		}

		// Удаляем рекомендацию после выполнения действия
		mu.Lock()
		delete(recStates, uid)
		mu.Unlock()

		// Ответ пользователю
		actionText := "Вы оценили профиль."
		if act == actionv1.Action_ACTION_LIKE {
			actionText = "Вы поставили лайк."
		} else if act == actionv1.Action_ACTION_DISLIKE {
			actionText = "Вы поставили дизлайк."
		}

		return c.Send(actionText, &telebot.ReplyMarkup{RemoveKeyboard: true})
	}
}

func listenMatches(bot *telebot.Bot, profClient profilev1.ProfileServiceClient, matchClient matchv1.MatchServiceClient) {
	stream, err := matchClient.ListenMatches(context.Background(), &matchv1.ListenMatchesRequest{})
	if err != nil {
		log.Printf("ListenMatches error: %v", err)
		return
	}
	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Printf("match stream closed: %v", err)
			return
		}
		go func(f, t int64) {
			pf, err1 := profClient.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: f})
			pt, err2 := profClient.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: t})
			if err1 != nil || err2 != nil {
				return
			}

			// Отправляем сообщение каждому пользователю с указанием другого пользователя в качестве матчевого партнера
			msgForF := fmt.Sprintf("У вас метч с %s", pt.Profile.FullName)
			msgForT := fmt.Sprintf("У вас метч с %s", pf.Profile.FullName)

			// Отправляем сообщения
			bot.Send(&telebot.User{ID: f}, msgForF)
			bot.Send(&telebot.User{ID: t}, msgForT)
		}(resp.FromUserId, resp.ToUserId)
	}
}
