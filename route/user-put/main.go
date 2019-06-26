package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chrisxue815/realworld-aws-lambda-dynamodb-go/model"
	"github.com/chrisxue815/realworld-aws-lambda-dynamodb-go/service"
	"github.com/chrisxue815/realworld-aws-lambda-dynamodb-go/util"
)

type RequestBody struct {
	User UserRequest `json:"user"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Image    string `json:"image"`
	Bio      string `json:"bio"`
}

type ResponseBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Image    string `json:"image"`
	Bio      string `json:"bio"`
	Token    string `json:"token"`
}

func Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	oldUser, token, err := service.GetCurrentUser(request.Headers["Authorization"])
	if err != nil {
		return util.NewUnauthorizedResponse()
	}

	requestBody := RequestBody{}
	err = json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return util.NewErrorResponse(err)
	}

	var password []byte
	if requestBody.User.Password != "" {
		const minPasswordLength = 0
		if len(requestBody.User.Password) < minPasswordLength {
			return util.NewErrorResponse(util.NewInputError("password", fmt.Sprintf("must be at least %d characters in length", minPasswordLength)))
		}

		password, err = service.Scrypt(requestBody.User.Password)
		if err != nil {
			return util.NewErrorResponse(err)
		}
	}

	newUser := model.User{
		Username: oldUser.Username,
		Email:    requestBody.User.Email,
		Password: password,
		Image:    requestBody.User.Image,
		Bio:      requestBody.User.Bio,
	}

	err = service.UpdateUser(*oldUser, newUser)
	if err != nil {
		return util.NewErrorResponse(err)
	}

	responseBody := ResponseBody{
		Username: newUser.Username,
		Email:    newUser.Email,
		Image:    newUser.Image,
		Bio:      newUser.Bio,
		Token:    token,
	}

	return util.NewSuccessResponse(200, responseBody)
}

func main() {
	lambda.Start(Handle)
}
