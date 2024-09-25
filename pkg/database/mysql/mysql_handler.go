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
	createUser                    *sql.Stmt
	checkUser                     *sql.Stmt
	getUserByEmail                *sql.Stmt
	getBySessionKey               *sql.Stmt
	getUserPortfolios             *sql.Stmt
	getUserTransactions           *sql.Stmt
	createNewTransaction          *sql.Stmt
	addNewForumPost               *sql.Stmt
	getSingleForumPost            *sql.Stmt
	getAllForums                  *sql.Stmt
	sendMessage                   *sql.Stmt
	addComment                    *sql.Stmt
	getCommentsByForum            *sql.Stmt
	createGroup                   *sql.Stmt
	addGroupMember                *sql.Stmt
	sendGroupMessage              *sql.Stmt
	getGroupMessages              *sql.Stmt
	getGroupAdmin                 *sql.Stmt
	checkIfMember                 *sql.Stmt
	getChats                      *sql.Stmt
	createConnectionRequest       *sql.Stmt
	getConnectionRequest          *sql.Stmt
	updateConnectionRequest       *sql.Stmt
	createConnection              *sql.Stmt
	getUserConnections            *sql.Stmt
	checkIfConnected              *sql.Stmt
	checkPendingConnection        *sql.Stmt
	requestGroupMembership        *sql.Stmt
	accept_declineGroupMembership *sql.Stmt
	fetchPendingMembershipRequest *sql.Stmt
	declineRequest                *sql.Stmt
	checkPendingMembershipRequest *sql.Stmt
}

func NewMySQLDatabase(db *sql.DB) (*mysqlDatabase, error) {
	var (
		createUser                     = "INSERT INTO users(username, password, email, degree, grad_year, current_job, phone, session_key, profile_picture, linkedin_profile, twitter_profile) VALUES(?,?,?,?,?,?,?,?,?,?,?);"
		checkUser                      = "SELECT * FROM users WHERE email = ? AND password=?;"
		getUserByEmail                 = "SELECT * FROM users WHERE email = ?;"
		getBySessionKey                = "SELECT * FROM users WHERE session_key=?;"
		getUserPortfolios              = "SELECT * FROM portfolio_orders WHERE user_email = ?;"
		getUserTransactions            = "SELECT * FROM transactions WHERE user_email = ?;"
		createNewTransaction           = "INSERT INTO transactions(from_user_id, from_user_email, to_user_id, to_user_email, type, created_at, updated_at, amount, user_email) VALUES(?,?,?,?,?,?,?,?,?);"
		addNewForumPost                = "INSERT INTO forums(title, description, author, slug, created_at, updated_at) VALUES(?,?,?,?,?,?);"
		getSingleForumPost             = "SELECT * FROM forums WHERE slug = ?;"
		getAllForums                   = "SELECT title, description, author, slug, created_at, updated_at FROM forums;"
		sendMessage                    = "INSERT INTO chat_messages(sender, recipient, message, created_at, updated_at) VALUES(?,?,?,?,?);"
		addComment                     = "INSERT INTO comments(user_id, forum_id, comment) VALUES(?, ?, ?);"
		getCommentsByForum             = "SELECT c.id, u.username, c.comment, c.created_at FROM comments c JOIN users u ON c.user_id = u.id WHERE c.forum_id = ? ORDER BY c.created_at ASC;"
		createGroup                    = "INSERT INTO groupie (name, created_by) VALUES (?,?);"
		addGroupMember                 = "INSERT INTO group_members (group_id, user_id) VALUES (?, ?);"
		sendGroupMessage               = "INSERT INTO group_messages (group_id, user_id, message) VALUES (?, ?, ?);"
		getGroupMessages               = "SELECT gm.id, u.username, gm.message, gm.created_at FROM group_messages gm JOIN users u ON gm.user_id = u.id WHERE gm.group_id = ? ORDER BY gm.created_at ASC;"
		getGroupAdmin                  = "SELECT u.id, u.username, u.email FROM groupie g JOIN users u ON g.created_by = u.id WHERE g.id = ?;"
		checkIfMember                  = "SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?;"
		getChats                       = "SELECT id, sender, recipient, message, created_at, updated_at FROM chat_messages WHERE (sender = ? AND recipient = ?) OR (sender = ? AND recipient = ?) ORDER BY created_at ASC;"
		createConnectionRequest        = "INSERT INTO connection_requests (from_id, to_id) VALUES (?, ?);"
		getConnectionRequest           = "SELECT id, from_id, to_id, status, created_at, updated_at FROM connection_requests WHERE from_id = ? AND to_id = ?;"
		updateConnectionRequest        = "UPDATE connection_requests SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;"
		createConnection               = "INSERT INTO connections (user_id, connection_user_id) VALUES (?, ?);"
		getUserConnections             = "SELECT u.id, u.username, u.email, u.grad_year, u.degree, u.current_job, u.phone, u.linkedin_profile, u.twitter_profile FROM connections c JOIN users u ON u.id = c.connection_user_id WHERE c.user_id = ?;"
		checkIfConnected               = "SELECT COUNT(*) FROM connections WHERE (user_id = ? AND connection_user_id = ?) OR (user_id = ? AND connection_user_id = ?);"
		checkPendingConnection         = "SELECT COUNT(*) FROM connection_requests WHERE from_id = ? AND to_id = ? AND status = 'pending';"
		requestGroupMembership         = "INSERT INTO group_membership_requests (group_id, user_id) VALUES (?, ?);"
		accept_decline_GroupMembership = "UPDATE group_membership_requests SET status = ? WHERE group_id = ? AND user_id = ?;"
		fetchPendingMembershipRequest  = "SELECT u.id, u.username, u.email, u.degree, u.grad_year, u.current_job, u.phone, u.profile_picture, u.linkedin_profile, u.twitter_profile, u.created_at, u.updated_at FROM group_membership_requests gmr JOIN users u ON gmr.user_id = u.id WHERE gmr.group_id = ? AND gmr.status = 'pending';"
		declineRequest                 = "DELETE FROM group_membership_requests WHERE group_id = ? AND user_id = ?;" //removes the user from the pending list
		checkPendingMemebershipRequest = "SELECT COUNT(*) FROM group_membership_requests WHERE group_id = ? AND user_id = ? AND status = 'pending';"
		database                       = &mysqlDatabase{}
		err                            error
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
	if database.createConnectionRequest, err = db.Prepare(createConnectionRequest); err != nil {
		return nil, err
	}
	if database.getConnectionRequest, err = db.Prepare(getConnectionRequest); err != nil {
		return nil, err
	}
	if database.updateConnectionRequest, err = db.Prepare(updateConnectionRequest); err != nil {
		return nil, err
	}
	if database.createConnection, err = db.Prepare(createConnection); err != nil {
		return nil, err
	}
	if database.getUserConnections, err = db.Prepare(getUserConnections); err != nil {
		return nil, err
	}
	if database.checkIfConnected, err = db.Prepare(checkIfConnected); err != nil {
		return nil, err
	}
	if database.checkPendingConnection, err = db.Prepare(checkPendingConnection); err != nil {
		return nil, err
	}
	if database.requestGroupMembership, err = db.Prepare(requestGroupMembership); err != nil {
		return nil, err
	}
	if database.accept_declineGroupMembership, err = db.Prepare(accept_decline_GroupMembership); err != nil {
		return nil, err
	}
	if database.fetchPendingMembershipRequest, err = db.Prepare(fetchPendingMembershipRequest); err != nil {
		return nil, err
	}
	if database.declineRequest, err = db.Prepare(declineRequest); err != nil {
		return nil, err
	}
	if database.checkPendingMembershipRequest, err = db.Prepare(checkPendingMemebershipRequest); err != nil {
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

func (db *mysqlDatabase) GetSingleForumPost(ctx context.Context, slug string) (*model.ForumPost, error) {
	forum := &model.ForumPost{}
	getForumBySlug := db.getSingleForumPost.QueryRowContext(ctx, slug)
	err := getForumBySlug.Scan(&forum.Id, &forum.Title, &forum.Description, &forum.Author, &forum.Slug, &forum.CreatedAt, &forum.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return forum, nil
}

func (db *mysqlDatabase) GetAllForums(ctx context.Context) (*[]model.ForumPost, error) {
	var forum = model.ForumPost{}
	var forums = []model.ForumPost{}
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
		log.Println("err fetching user chats:", err)
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

func (db *mysqlDatabase) CreateConnectionRequest(ctx context.Context, fromUserId, toUserId int) (bool, error) {
	con_req, err := db.createConnectionRequest.ExecContext(ctx, fromUserId, toUserId)
	if err != nil {
		log.Println("err creating connection request:", err)
		return false, err
	}
	con_req_lid, err := con_req.LastInsertId()
	if err != nil {
		return false, err
	}
	if con_req_lid <= 0 {
		log.Printf("last insert id: %d", con_req_lid)
		return false, fmt.Errorf("unable to create new connection request")
	}
	return true, nil
}

func (db *mysqlDatabase) GetConnectionRequest(ctx context.Context, fromUserId, toUserId int) (*model.ConnectionRequest, error) {
	req := &model.ConnectionRequest{}
	err := db.getConnectionRequest.QueryRowContext(ctx, fromUserId, toUserId).Scan(&req.Id, &req.FromUserId, &req.ToUserId, &req.Status, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		log.Println("err fetching connection request:", err)
		return nil, err
	}
	return req, nil
}

func (db *mysqlDatabase) UpdateConnectionRequest(ctx context.Context, reqId int, status string) (bool, error) {
	con_req, err := db.updateConnectionRequest.ExecContext(ctx, status, reqId)
	if err != nil {
		log.Println("err updating connection request:", err)
		return false, err
	}
	con_req_rwa, err := con_req.RowsAffected()
	if err != nil {
		return false, err
	}
	if con_req_rwa <= 0 {
		return false, fmt.Errorf("err updating connection request")
	}
	return true, nil
}

func (db *mysqlDatabase) CreateConnection(ctx context.Context, userId, connectionUserId int) (bool, error) {
	con_req, err := db.createConnection.ExecContext(ctx, userId, connectionUserId)
	if err != nil {
		log.Println("Error creating connection:", err)
		return false, err
	}
	con_req_lid, err := con_req.LastInsertId()
	if err != nil {
		return false, err
	}
	if con_req_lid <= 0 {
		log.Printf("last insert id: %d", con_req_lid)
		return false, fmt.Errorf("unable to create new connection request")
	}
	return true, nil
}

func (db *mysqlDatabase) GetUserConnections(ctx context.Context, userId int) ([]*model.User, error) {
	rows, err := db.getUserConnections.QueryContext(ctx, userId)
	if err != nil {
		log.Println("Error fetching user connections:", err)
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.Id, &user.Username, &user.Email, &user.GradYear, &user.Degree, &user.CurrentJob, &user.Phone, &user.LinkedinProfile, &user.TwitterProfile); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (db *mysqlDatabase) CheckIfConnected(ctx context.Context, userID1, userID2 int) (bool, error) {
	var count int
	err := db.checkIfConnected.QueryRowContext(ctx, userID1, userID2, userID2, userID1).Scan(&count)
	if err != nil {
		log.Println("error checking existing connection:", err)
		return false, err
	}
	return count > 0, nil
}

func (db *mysqlDatabase) CheckPendingConnection(ctx context.Context, fromUserID, toUserID int) (bool, error) {
	var count int
	err := db.checkPendingConnection.QueryRowContext(ctx, fromUserID, toUserID).Scan(&count)
	if err != nil {
		log.Println("Error checking pending request:", err)
		return false, err
	}
	return count > 0, nil
}

func (db *mysqlDatabase) RequestGroupMembership(ctx context.Context, group_id, user_id int) (bool, error) {
	requst, err := db.requestGroupMembership.ExecContext(ctx, group_id, user_id)
	if err != nil {
		log.Printf("%v", err)
		return false, err
	}
	req_lid, err := requst.LastInsertId()
	if err != nil {
		return false, err
	}
	if req_lid <= 0 {
		return false, err
	}
	return true, err
}

func (db *mysqlDatabase) UpdateGroupMembershipRequest(ctx context.Context, status string, group_id int, user_id int) (bool, error) {
	accept, err := db.accept_declineGroupMembership.ExecContext(ctx, status, group_id, user_id)
	if err != nil {
		log.Println("here here")
		log.Printf("%v", err)
		return false, err
	}
	a_lid, err := accept.RowsAffected()
	if err != nil {
		return false, err
	}
	if a_lid <= 0 {
		log.Println("or here")
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) FetchPendingMembershipRequest(ctx context.Context, group_id int) ([]*model.User, error) {
	var pendingRequests []*model.User
	pendings, err := db.fetchPendingMembershipRequest.QueryContext(ctx, group_id)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	defer pendings.Close()
	for pendings.Next() {
		var pendingRequest model.User
		err := pendings.Scan(&pendingRequest.Id, &pendingRequest.Username, &pendingRequest.Email, &pendingRequest.Degree, &pendingRequest.GradYear, &pendingRequest.CurrentJob, &pendingRequest.Phone, &pendingRequest.ProfilePicture, &pendingRequest.LinkedinProfile, &pendingRequest.TwitterProfile, &pendingRequest.CreatedAt, &pendingRequest.UpdatedAt)
		if err != nil {
			log.Printf("%v", err)
			return nil, err
		}
		pendingRequests = append(pendingRequests, &pendingRequest)
	}
	return pendingRequests, nil
}

func (db *mysqlDatabase) DeclineMembershipRequest(ctx context.Context, groupID, userID int) (bool, error) {
	res, err := db.declineRequest.ExecContext(ctx, groupID, userID)
	if err != nil {
		return false, err
	}
	res_rwa, err := res.RowsAffected()
	if err != nil {
		log.Printf("err: %v", err)
		return false, err
	}
	if res_rwa <= 0 {
		log.Printf("rws affected: %d", res_rwa)
		return false, err
	}
	return true, nil
}

func (db *mysqlDatabase) CheckPendingMembershipRequest(ctx context.Context, groupID, userID int) (bool, error) {
	var count int
	err := db.checkPendingMembershipRequest.QueryRowContext(ctx, groupID, userID).Scan(&count)
	if err != nil {
		log.Printf("err checking pending membership request:%v", err)
		return false, err
	}
	return count > 0, nil
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
	db.createConnection.Close()
	db.createConnectionRequest.Close()
	db.getConnectionRequest.Close()
	db.getUserConnections.Close()
	db.updateConnectionRequest.Close()
	db.checkIfConnected.Close()
	db.checkPendingConnection.Close()
	db.requestGroupMembership.Close()
	db.accept_declineGroupMembership.Close()
	db.fetchPendingMembershipRequest.Close()
	db.declineRequest.Close()
	db.checkPendingMembershipRequest.Close()
	return nil
}
