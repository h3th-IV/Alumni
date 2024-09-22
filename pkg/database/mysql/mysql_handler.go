package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jim-nnamdi/jinx/pkg/model"
)

var (
	_ Database  = &mysqlDatabase{}
	_ io.Closer = &mysqlDatabase{}
)

type mysqlDatabase struct {
	createUser           *sql.Stmt
	checkUser            *sql.Stmt
	getUserByEmail       *sql.Stmt
	getBySessionKey      *sql.Stmt
	getUserPortfolios    *sql.Stmt
	getUserTransactions  *sql.Stmt
	createNewTransaction *sql.Stmt
	addNewForumPost      *sql.Stmt
	getSingleForumPost   *sql.Stmt
	getAllForums         *sql.Stmt
	sendMessage          *sql.Stmt
	addComment           *sql.Stmt
	getCommentsByForum   *sql.Stmt
	createGroup          *sql.Stmt
	addGroupMember       *sql.Stmt
	sendGroupMessage     *sql.Stmt
	getGroupMessages     *sql.Stmt
	getGroupAdmin        *sql.Stmt
	checkIfMember        *sql.Stmt
	getChats             *sql.Stmt
}

func NewMySQLDatabase(db *sql.DB) (*mysqlDatabase, error) {
	var (
		createUser           = "INSERT INTO users(username, password, email, degree, grad_year,current_job, phone, session_key,profile_picture,linkedin_profile,twitter_profile) VALUES(?,?,?,?,?,?,?,?,?,?,?);"
		checkUser            = "SELECT * FROM users where email = ? AND password=?;"
		getUserByEmail       = "SELECT * FROM users where email = ?;"
		getBySessionKey      = "SELECT * FROM users where session_key=?;"
		getUserPortfolios    = "SELECT * FROM portfolio_orders WHERE `user_email` = ?;"
		getUserTransactions  = "SELECT * FROM transactions WHERE `user_email` = ?;"
		createNewTransaction = "INSERT INTO transactions(from_user_id,from_user_email, to_user_id, to_user_email,type,created_at,updated_at,amount,user_email) VALUES(?,?,?,?,?,?,?,?,?);"
		addNewForumPost      = "INSERT INTO forums(title, description, author, slug, created_at, updated_at) VALUES (?,?,?,?,?,?)"
		getSingleForumPost   = "SELECT * FROM forums WHERE `slug` = ?;"
		getAllForums         = "SELECT title, description, author, slug, created_at, updated_at FROM forums"
		sendMessage          = "INSERT INTO chat_messages (sender, recipient, message, created_at,updated_at) VALUES (?,?,?,?,?)"
		addComment           = "INSERT INTO comments (user_id, forum_id, comment) VALUES (?, ?, ?)"
		getCommentsByForum   = "SELECT c.id, u.username, c.comment, c.created_at FROM comments c JOIN users u ON c.user_id = u.id WHERE c.forum_id = ? ORDER BY c.created_at ASC"
		createGroup          = "INSERT INTO groups (name, created_by) VALUES (?,?)"
		addGroupMember       = "INSERT INTO group_members (group_id, user_id) VALUES (?, ?)"
		sendGroupMessage     = "INSERT INTO group_messages (group_id, user_id, message) VALUES (?, ?, ?)"
		getGroupMessages     = "SELECT gm.id, u.username, gm.message, gm.created_at FROM group_messages gm JOIN users u ON gm.user_id = u.id WHERE gm.group_id = ? ORDER BY gm.created_at ASC"
		getGroupAdmin        = "SELECT u.id, u.username, u.email FROM groups g JOIN users u ON g.created_by = u.id WHERE g.id = ?"
		checkIfMember        = "SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?"
		getChats             = "SELECT id, sender, recipient, message, created_at, updated_at FROM chat_messages WHERE (sender = ? AND recipient = ?) OR (sender = ? AND recipient = ?) ORDER BY created_at ASC"
		database             = &mysqlDatabase{}
		err                  error
	)
	if database.createUser, err = db.Prepare(createUser); err != nil {
		return nil, err
	}
	if database.checkUser, err = db.Prepare(checkUser); err != nil {
		return nil, err
	}
	if database.getUserByEmail, err = db.Prepare(getUserByEmail); err != nil {
		return nil, err
	}
	if database.getBySessionKey, err = db.Prepare(getBySessionKey); err != nil {
		return nil, err
	}
	if database.getUserPortfolios, err = db.Prepare(getUserPortfolios); err != nil {
		return nil, err
	}
	if database.getUserTransactions, err = db.Prepare(getUserTransactions); err != nil {
		return nil, err
	}
	if database.createNewTransaction, err = db.Prepare(createNewTransaction); err != nil {
		return nil, err
	}
	if database.addNewForumPost, err = db.Prepare(addNewForumPost); err != nil {
		return nil, err
	}
	if database.getSingleForumPost, err = db.Prepare(getSingleForumPost); err != nil {
		return nil, err
	}
	if database.getAllForums, err = db.Prepare(getAllForums); err != nil {
		return nil, err
	}
	if database.sendMessage, err = db.Prepare(sendMessage); err != nil {
		return nil, err
	}
	if database.addComment, err = db.Prepare(addComment); err != nil {
		return nil, err
	}
	if database.getCommentsByForum, err = db.Prepare(getCommentsByForum); err != nil {
		return nil, err
	}
	if database.createGroup, err = db.Prepare(createGroup); err != nil {
		return nil, err
	}
	if database.addGroupMember, err = db.Prepare(addGroupMember); err != nil {
		return nil, err
	}
	if database.sendGroupMessage, err = db.Prepare(sendGroupMessage); err != nil {
		return nil, err
	}
	if database.getGroupMessages, err = db.Prepare(getGroupMessages); err != nil {
		return nil, err
	}
	if database.getGroupAdmin, err = db.Prepare(getGroupAdmin); err != nil {
		return nil, err
	}
	if database.checkIfMember, err = db.Prepare(checkIfMember); err != nil {
		return nil, err
	}
	if database.getChats, err = db.Prepare(getChats); err != nil {
		return nil, err
	}
	return database, nil
}

func (db *mysqlDatabase) CreateUser(ctx context.Context, username string, password string, email string, degree string, gradyear string, currentjob string, phone string, sessionkey string, profilepicture string, linkedinprofile string, twitterprofile string) (bool, error) {
	userQuery, err := db.createUser.ExecContext(ctx, username, password, email, degree, gradyear, currentjob, phone, sessionkey, profilepicture, linkedinprofile, twitterprofile)
	if err != nil {
		return false, err
	}
	lastId, err := userQuery.LastInsertId()
	if err != nil {
		return false, err
	}
	if lastId == 0 || lastId < 1 {
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	getUserByEmail := db.getUserByEmail.QueryRowContext(ctx, email)
	err := getUserByEmail.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Degree, &user.GradYear, &user.CurrentJob, &user.Phone, &user.SessionKey, &user.ProfilePicture, &user.LinkedinProfile, &user.TwitterProfile, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Println("get user by email", err)
		return nil, err
	}
	return user, nil
}

func (db *mysqlDatabase) CheckUser(ctx context.Context, email string, password string) (*model.User, error) {
	user := &model.User{}
	getUserByEmail := db.checkUser.QueryRowContext(ctx, email, password)
	err := getUserByEmail.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Degree, &user.GradYear, &user.CurrentJob, &user.Phone, &user.SessionKey, &user.ProfilePicture, &user.LinkedinProfile, &user.TwitterProfile, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Println("checkuser", err)
		return nil, err
	}
	return user, nil
}

func (db *mysqlDatabase) GetBySessionKey(ctx context.Context, sessionkey string) (*model.User, error) {
	user := &model.User{}
	getBySessionKey := db.getBySessionKey.QueryRowContext(ctx, sessionkey)
	err := getBySessionKey.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Degree, &user.GradYear, &user.CurrentJob, &user.Phone, &user.SessionKey, &user.ProfilePicture, &user.LinkedinProfile, &user.TwitterProfile, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *mysqlDatabase) GetUserPortfolio(ctx context.Context, user_email string) (*[]model.PortfolioOrder, error) {
	var portfolioOrders = []model.PortfolioOrder{}
	var portfolioOrder = model.PortfolioOrder{}
	getPortfolio, err := db.getUserPortfolios.QueryContext(ctx, user_email)
	if err != nil {
		return nil, err
	}
	defer getPortfolio.Close()
	for getPortfolio.Next() {
		err := getPortfolio.Scan(&portfolioOrder.Id, &portfolioOrder.Type, &portfolioOrder.Security, &portfolioOrder.Unit, &portfolioOrder.Status, &portfolioOrder.Cancelled, &portfolioOrder.UserID, &portfolioOrder.UserEmail, &portfolioOrder.CreatedAt, &portfolioOrder.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}
	portfolioOrders = append(portfolioOrders, portfolioOrder)
	return &portfolioOrders, nil
}

func (db *mysqlDatabase) GetUserTransactions(ctx context.Context, user_email string) (*[]model.Transaction, error) {
	var transaction = model.Transaction{}
	var transactions = []model.Transaction{}
	getTransactions, err := db.getUserTransactions.QueryContext(ctx, user_email)
	if err != nil {
		return nil, err
	}
	defer getTransactions.Close()
	for getTransactions.Next() {
		err := getTransactions.Scan(&transaction.Id, &transaction.FromUserID, &transaction.FromUserEmail, &transaction.ToUserID, &transaction.ToUserEmail, &transaction.TransactionType, &transaction.CreatedAt, &transaction.UpdatedAt, &transaction.Amount, &transaction.UserEmail)
		if err != nil {
			return nil, err
		}
	}
	transactions = append(transactions, transaction)
	return &transactions, nil
}

func (db *mysqlDatabase) CreateNewTransaction(ctx context.Context, from_user int, from_user_email string, to_user int, to_user_email string, transactiontype string, created_at time.Time, updated_at time.Time, amount int, user_email string) (bool, error) {
	createNewTx, err := db.createNewTransaction.ExecContext(ctx, from_user, from_user_email, to_user, to_user_email, transactiontype, created_at, updated_at, amount, user_email)
	if err != nil {
		return false, err
	}
	lastInsert, err := createNewTx.LastInsertId()
	if err != nil {
		return false, err
	}
	if lastInsert <= 0 {
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) AddNewForumPost(ctx context.Context, title string, description string, author string, slug string, created_at time.Time, updated_at time.Time) (bool, error) {
	createNewForum, err := db.addNewForumPost.ExecContext(ctx, title, description, author, slug, created_at, updated_at)
	if err != nil {
		return false, err
	}
	lastInsert, err := createNewForum.LastInsertId()
	if err != nil {
		return false, err
	}
	if lastInsert <= 0 {
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) SendMessage(ctx context.Context, senderId int, receiverId int, message string, createdAt time.Time, updatedAt time.Time) (bool, error) {
	sendmessage, err := db.sendMessage.ExecContext(ctx, senderId, receiverId, message, createdAt, updatedAt)
	if err != nil {
		return false, err
	}
	lastInsert, err := sendmessage.LastInsertId()
	if err != nil {
		return false, err
	}
	if lastInsert <= 0 {
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) GetSingleForumPost(ctx context.Context, slug string) (*model.Forum, error) {
	forum := &model.Forum{}
	getForumBySlug := db.getSingleForumPost.QueryRowContext(ctx, slug)
	err := getForumBySlug.Scan(&forum.Id, &forum.Title, &forum.Description, &forum.Author, &forum.Slug, &forum.CreatedAt, &forum.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return forum, nil
}

func (db *mysqlDatabase) GetAllForums(ctx context.Context) (*[]model.Forum, error) {
	var forum = model.Forum{}
	var forums = []model.Forum{}
	getForums, err := db.getAllForums.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	for getForums.Next() {
		err := getForums.Scan(&forum.Title, &forum.Description, &forum.Author, &forum.Slug, &forum.CreatedAt, &forum.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}
	forums = append(forums, forum)
	return &forums, nil
}

func (db *mysqlDatabase) AddComment(ctx context.Context, userID int, forumID int, comment string) (bool, error) {
	incoming, err := db.addComment.ExecContext(ctx, userID, forumID, comment)
	if err != nil {
		return false, err
	}
	c_lid, err := incoming.LastInsertId()
	if err != nil {
		return false, err
	}
	if c_lid <= 0 {
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) GetCommentsByForumID(ctx context.Context, forumID int) ([]model.Comment, error) {
	comments := []model.Comment{}
	rows, err := db.getCommentsByForum.QueryContext(ctx, forumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(&comment.ID, &comment.Username, &comment.Comment, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

// CreateGroup creates a new group.
func (db *mysqlDatabase) CreateGroup(ctx context.Context, name string, createdby int) (int, error) {
	result, err := db.createGroup.ExecContext(ctx, name, createdby)
	if err != nil {
		return 0, err
	}
	g_lid, err := result.LastInsertId()
	if err != nil {
		log.Printf("last insert id: %v", g_lid)
		return 0, err
	}
	if g_lid <= 0 {
		return 0, fmt.Errorf("unable to create new group")
	}
	return int(g_lid), nil
}

// AddGroupMember adds a new member to the group.
func (db *mysqlDatabase) AddGroupMember(ctx context.Context, groupID int, userID int) (bool, error) {
	group, err := db.addGroupMember.ExecContext(ctx, groupID, userID)
	if err != nil {
		return false, err
	}
	mem_lid, err := group.LastInsertId()
	if err != nil {
		log.Printf("last insert id: %v", mem_lid)
		return false, err
	}
	if mem_lid <= 0 {
		return false, fmt.Errorf("unable to add new member")
	}
	return true, nil
}

func (db *mysqlDatabase) SendGroupMessage(ctx context.Context, groupID int, userID int, message string) (bool, error) {
	newMessage, err := db.sendGroupMessage.ExecContext(ctx, groupID, userID, message)
	if err != nil {
		return false, err
	}
	nm_lid, err := newMessage.LastInsertId()
	if err != nil {
		log.Printf("last insert id: %v", nm_lid)
		return false, err
	}
	if nm_lid <= 0 {
		return false, fmt.Errorf("unable to send group message")
	}
	return true, nil
}

func (db *mysqlDatabase) GetGroupMessages(ctx context.Context, groupID int) ([]model.GroupMessage, error) {
	messages := []model.GroupMessage{}
	rows, err := db.getGroupMessages.QueryContext(ctx, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message model.GroupMessage
		if err := rows.Scan(&message.ID, &message.Username, &message.Message, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (db *mysqlDatabase) GetGroupCreator(ctx context.Context, groupID int) (*model.User, error) {
	user := &model.User{}

	// Execute the query
	row := db.getGroupAdmin.QueryRowContext(ctx, groupID)
	err := row.Scan(&user.Id, &user.Username, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no group found with id %d", groupID)
		}
		log.Println("Error fetching group creator:", err)
		return nil, err
	}

	return user, nil
}

// CheckGroupMembership checks if a user is a member of a specific group.
func (db *mysqlDatabase) CheckGroupMembership(ctx context.Context, groupID int, userID int) (bool, error) {
	var count int
	err := db.checkIfMember.QueryRowContext(ctx, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *mysqlDatabase) FetchUserChats(ctx context.Context, userID1, userID2 int) ([]*model.Chat, error) {

	rows, err := db.getChats.QueryContext(ctx, userID1, userID2, userID2, userID1)
	if err != nil {
		log.Println("Error fetching user chats:", err)
		return nil, err
	}
	defer rows.Close()

	var chats []*model.Chat
	for rows.Next() {
		var chat model.Chat
		err := rows.Scan(&chat.Id, &chat.SenderID, &chat.RecipientID, &chat.Message, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	return chats, nil
}

func (db *mysqlDatabase) Close() error {
	db.createUser.Close()
	db.checkUser.Close()
	db.getBySessionKey.Close()
	db.getUserByEmail.Close()
	db.getUserPortfolios.Close()
	db.getUserTransactions.Close()
	db.createNewTransaction.Close()
	db.addNewForumPost.Close()
	db.getSingleForumPost.Close()
	db.addComment.Close()
	db.sendMessage.Close()
	db.getCommentsByForum.Close()
	db.createGroup.Close()
	db.addGroupMember.Close()
	db.sendGroupMessage.Close()
	db.getGroupAdmin.Close()
	db.checkIfMember.Close()
	db.getChats.Close()
	return nil
}
