package mysql

import (
	"context"
	"io"
	"time"

	"github.com/jim-nnamdi/jinx/pkg/model"
)

//go:generate mockgen -destination=mocks/mock_database.go -package=mocks

type Database interface {
	io.Closer

	/* user interaction queries */
	CreateUser(ctx context.Context, username string, password string, email string, degree string, gradyear string, currentjob string, phone string, sessionkey string, profilepicture string, linkedinprofile string, twitterprofile string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CheckUser(ctx context.Context, email string, password string) (*model.User, error)
	GetBySessionKey(ctx context.Context, sessionkey string) (*model.User, error)
	GetUserPortfolio(ctx context.Context, user_email string) (*[]model.PortfolioOrder, error)

	/* transactions */
	GetUserTransactions(ctx context.Context, user_email string) (*[]model.Transaction, error)
	CreateNewTransaction(ctx context.Context, from_user int, from_user_email string, to_user int, to_user_email string, transactiontype string, created_at time.Time, updated_at time.Time, amount int, user_email string) (bool, error)

	/*forum and messages*/
	AddNewForumPost(ctx context.Context, title string, description string, author string, slug string, created_at time.Time, updated_at time.Time) (bool, error)
	GetSingleForumPost(ctx context.Context, slug string) (*model.Forum, error)
	GetAllForums(ctx context.Context) (*[]model.Forum, error)
	GetCommentsByForumID(ctx context.Context, forumID int) ([]model.Comment, error)
	SendMessage(ctx context.Context, senderId int, receiverId int, message string, createdAt time.Time, updatedAt time.Time) (bool, error)
	AddComment(ctx context.Context, userID int, forumID int, comment string) (bool, error)
	CreateGroup(ctx context.Context, name string, userID int) (int, error)
	AddGroupMember(ctx context.Context, groupID int, userID int) (bool, error)
	SendGroupMessage(ctx context.Context, groupID int, userID int, message string) (bool, error)
	GetGroupMessages(ctx context.Context, groupID int) ([]model.GroupMessage, error)
	GetGroupCreator(ctx context.Context, groupID int) (*model.User, error)
	CheckGroupMembership(ctx context.Context, groupID int, userID int) (bool, error)
	FetchUserChats(ctx context.Context, userID1, userID2 int) ([]*model.Chat, error)
}
