/set-profile
{
	"age": 19,
	"full_name": "Tarasov Mihail",
	"gender": "GENDER_MALE",
	"user_id": 0
}

{
	"age": 18,
	"description": "super villain",
	"full_name": "Cat woman",
	"gender": "GENDER_FEMALE",
	"photo": "adfsaf",
	"user_id": 1
}

{
	"age": 25,
	"description": "super woman",
	"full_name": "Lara Croft",
	"gender": "GENDER_FEMALE",
	"user_id": 2
}

{
	"age": 45,
	"full_name": "Mary Cury",
	"gender": "GENDER_FEMALE",
	"user_id": 3
}

/set-filter
{
	"gender": "GENDER_FEMALE",
	"max_age": 50,
	"min_age": 15,
	"user_id": 0
}

/get-recommendation (should be cat woman)
{
	"user_id": 0
}

/action
{
	"action": "ACTION_LIKE",
	"from_user_id": 0,
	"to_user_id": 1
}

/get-recommendation (should be lara)
{
	"user_id": 0
}

/action
{
	"action": "ACTION_DISLIKE",
	"from_user_id": 0,
	"to_user_id": 2
}

/get-recommendation (should be cury)
{
	"user_id": 0
}

/action
{
	"action": "ACTION_LIKE",
	"from_user_id": 0,
	"to_user_id": 3
}

/listen-match

/set-filter
{
	"gender": "GENDER_MALE",
	"max_age": 50,
	"min_age": 15,
	"user_id": 1
}

/get-recommendation (should be Misha Tarasov)
{
	"user_id": 1
}

/action
{
	"action": "ACTION_LIKE",
	"from_user_id": 1,
	"to_user_id": 0
}
