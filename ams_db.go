package ams

import (
	"database/sql"
	"errors"
	"fmt"
)

var dbConn *sql.DB

//AppConnector ams对数据库操作的上下文
type AppConnector struct {
	dbConn *sql.DB
}

//数据库连接
var appConnector *AppConnector

var createDatabaseSQL string = "CREATE DATABASE ams;"

var createDomainSQL string = `CREATE TABLE ams_domain (
	id serial PRIMARY KEY NOT NULL,
	name varchar(128) NOT NULL,
	tenant_id int NOT NULL,
	company text,
	enable boolean NOT NULL,
	update timestamp(6) DEFAULT now()
 	);`

var createGroupSQL string = `CREATE TABLE ams_group (
	id serial PRIMARY KEY NOT NULL,
	name varchar(64) NOT NULL,
	group_desc text,
	parent_id int NULL,
	domain_id int NOT NULL,
	update timestamp(6) DEFAULT now()
	);`

var createUserSQL string = `CREATE TABLE ams_user (
	id serial PRIMARY KEY NOT NULL,
	username varchar(64) NOT NULL,
	password varchar(64) NOT NULL,
	group_id int NOT NULL,
	update timestamp(6) DEFAULT now()
	);`

var insertDomainSQL string = "INSERT INTO ams_domain(name,tenant_id,company,enable) VALUES('%s','%d','%s',%s) RETURNING id"
var insertGroupSQL string = "INSERT INTO ams_group(name,group_desc,parent_id,domain_id) VALUES('%s','%s','%d','%d') RETURNING id"
var insertUserSQL string = "INSERT INTO ams_user(username,password,group_id) VALUES('%s','%s','%d') RETURNING id"

var selectDomainSQL string = "SELECT id,name,tenant_id,company,enable FROM ams_domain where id=%d"
var selectGroupSQL string = "SELECT id,name,group_desc,parent_id,domain_id FROM ams_group where id=%d"
var selectUserSQL string = "SELECT id,username,password,group_id FROM ams_user where id=%d"

var selectEnabledUsersSQL string = "select u.id,u.username,u.password,u.group_id from ams_user as u left join ams_group as g on u.group_id=g.id where g.domain_id=%d"
var selectEnabledDomainsSQL string = "SELECT id,name,tenant_id,company FROM ams_domain where enable=true"

//查找domain下的username
var selectUsernameFromDomainSQL string = "select u.username from ams_user as u INNER join ams_group as g on u.group_id=g.id where g.domain_id=%d and u.username='%s' LIMIT 1"

var selectDomainIDByNameSQL string = "SELECT id FROM ams_domain WHERE name='%s'"
var selectGroupIDByNameSQL string = "SELECT id FROM ams_group WHERE name='%s'"
var selectUserIDByNameSQL string = "SELECT id FROM ams_user WHERE username='%s'"

var selectRealmByIDSQL string = "SELECT name FROM ams_domain WHERE id=%d"
var selectGroupNameByIDSQL string = "SELECT name FROM ams_group WHERE id=%d"
var selectUserNameByIDSQL string = "SELECT name FROM ams_user WHERE id=%d"

var selectOneGroupIDByDomainIDSQL string = "SELECT id from ams_group where domain_id=%d LIMIT 1"
var selectOneGroupIDByParentIDSQL string = "SELECT id FROM ams_group WHERE parent_id=%d LIMIT 1"
var selectOneUserIDByGroupIDSQL string = "SELECT id FROM ams_user WHERE group_id=%d LIMIT 1"

/*
var updateDomainSQL string = "UPDATE ams_domain SET name='%s',tenant_id=%d,company='%s',enable=%t where id=%d"
var updateGroupSQL string = "UPDATE ams_group SET name='%s',group_desc='%s',parent_id=%d,domain_id=%d where id=%d"
var updateUserSQL string = "UPDATE ams_user SET username='%s',password='%s',group_id=%d,status='%s',state='%s' where id=%d"
*/
var updateDomainSQL string = "UPDATE ams_domain SET name=$1,tenant_id=$2,company=$3,enable=$4 where id=$5"
var updateGroupSQL string = "UPDATE ams_group SET name=$1,group_desc=$2,parent_id=$3,domain_id=$4 where id=$5"
var updateUserSQL string = "UPDATE ams_user SET username=$1,password=$2,group_id=$3 where id=$4"

var deleteDomainSQL string = "DELETE FROM ams_domain WHERE id = $1"
var deleteGroupSQL string = "DELETE FROM ams_group WHERE id = $1"
var deleteUserSQL string = "DELETE FROM ams_user WHERE id = $1"

//GetDBConnector 返回与数据库的连接
func GetDBConnector() (*AppConnector, error) {
	if appConnector != nil {
		return appConnector, nil
	}
	return nil, errors.New("ams is not connected to the database")
}

//OpenDBConnector 建立与DB的连接并返回
func OpenDBConnector(host string, port string, user string, password string, dbName string) (*AppConnector, error) {
	if appConnector != nil {
		return nil, errors.New("ams already connected to the database")
	}
	dbSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	//dbSource := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&dbname=%s", user, password, addr, user, dbName)
	dbConn, err := sql.Open("postgres", dbSource)
	if err != nil {
		return nil, err
	}
	err = dbConn.Ping()
	if err != nil {
		return nil, err
	}
	Info.Println("Successfully connected database!")
	appConnector = new(AppConnector)
	appConnector.dbConn = dbConn
	return appConnector, nil
}

//CloseDBConnector 关闭DB连接
func (c *AppConnector) CloseDBConnector() {
	c.dbConn.Close()
	c = nil
}

//execSQL 执行sql create update delete语句
func (c *AppConnector) execSQL(SQL string, args ...interface{}) error {
	stmt, err := c.dbConn.Prepare(SQL)
	if err != nil {
		//Error.Println("execSQL Prepare ", err)
		return err
	}
	result, err := stmt.Exec(args...)
	if err != nil {
		//Error.Println("execSQL stmt.exec ", err)
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		//Error.Println("execSQL RowsAffected ", err)
		return err
	}
	return nil
}

//insertSQL 执行sql插入语句 获取自增id
func (c *AppConnector) insertSQL(SQL string, args ...interface{}) (int, error) {
	var id int
	insertSQL := fmt.Sprintf(SQL, args...)
	err := c.dbConn.QueryRow(insertSQL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

//CreateTable 创建表
func (c *AppConnector) CreateTable() {
	err := c.execSQL(createDomainSQL)
	if err != nil {
		Error.Println("create ams_domain fail ", err)
	}
	err = c.execSQL(createGroupSQL)
	if err != nil {
		Error.Println("create ams_group fail ", err)
	}
	err = c.execSQL(createUserSQL)
	if err != nil {
		Error.Println("create ams_user fail ", err)
	}
}

//InsertDomain 插入一条域数据
func (c *AppConnector) InsertDomain(name string, tenantID int, company string, enable string) (int, error) {
	id, err := c.insertSQL(insertDomainSQL, name, tenantID, company, enable)
	if err != nil {
		Error.Println(err)
		return 0, err
	}
	return id, nil
}

//InsertGroup 插入一条组数据
func (c *AppConnector) InsertGroup(name string, desc string, parentID int, domainID int) (int, error) {
	id, err := c.insertSQL(insertGroupSQL, name, desc, parentID, domainID)
	if err != nil {
		Error.Println(err)
		return 0, err
	}
	return id, nil
}

//InsertUser 插入一条号码数据
func (c *AppConnector) InsertUser(user string, password string, groupID int) (int, error) {
	id, err := c.insertSQL(insertUserSQL, user, password, groupID)
	if err != nil {
		Error.Println(err)
		return 0, err
	}
	return id, nil
}

//ReadDomain 查询domain信息
func (c *AppConnector) ReadDomain(id int) (*DomainInfo, error) {
	if c == nil {
		Error.Println("db context is null")
		return nil, errors.New("db context is null")
	}
	queryStr := fmt.Sprintf(selectDomainSQL, id)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	p := new(DomainInfo)
	for rows.Next() {
		err := rows.Scan(&p.id, &p.Name, &p.TenantID, &p.Company, &p.Enable)
		if err != nil {
			Error.Println(err)
		}
		// Debug.Println(p.id, p.Name, p.TenantID, p.Company, p.Enable)
	}
	return p, nil
}

//ReadGroup 查询group信息
func (c *AppConnector) ReadGroup(id int) (*GroupInfo, error) {
	if id <= 0 {
		return nil, errors.New("input parameters null")
	}
	queryStr := fmt.Sprintf(selectGroupSQL, id)

	rows, err := c.dbConn.Query(queryStr)
	if err != nil {
		Error.Println(err)
		return nil, err
	}
	defer rows.Close()
	p := new(GroupInfo)
	for rows.Next() {
		err := rows.Scan(&p.id, &p.Name, &p.GroupDesc, &p.ParentID, &p.DomainID) //id,name,desc,parent,domain
		if err != nil {
			Error.Println(err)
		}
		// Debug.Println(p.id, p.Name, p.GroupDesc, p.ParentID, p.DomainID)
	}
	return p, nil
}

//ReadUser 查询user信息
func (c *AppConnector) ReadUser(id int) (*UserInfo, error) {
	queryStr := fmt.Sprintf(selectUserSQL, id)

	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("ReadUser db query ", err)
		return nil, err
	}
	defer rows.Close()
	p := new(UserInfo)
	for rows.Next() {
		err := rows.Scan(&p.id, &p.Username, &p.Password, &p.GroupID)
		if err != nil {
			Error.Println(err)
		}
		// Debug.Println(p.id, p.Username, p.Password, p.GroupID)
	}
	return p, nil
}

//UpdateDomain 更新域信息
func (c *AppConnector) UpdateDomain(p *DomainInfo) error {
	err := c.execSQL(updateDomainSQL, p.Name, p.TenantID, p.Company, p.Enable, p.id)
	if err != nil {
		Error.Println(err)
		return err
	}
	return nil
}

//UpdateGroup 更新组信息
func (c *AppConnector) UpdateGroup(p *GroupInfo) error {
	err := c.execSQL(updateGroupSQL, p.Name, p.GroupDesc, p.ParentID, p.DomainID, p.id)
	if err != nil {
		Error.Println(err)
		return err
	}
	return nil
}

//UpdateUser 更新用户信息
func (c *AppConnector) UpdateUser(p *UserInfo) error {
	err := c.execSQL(updateUserSQL, p.Username, p.Password, p.GroupID, p.id)
	if err != nil {
		Error.Println(err)
		return err
	}
	return nil
}

//DeleteDomain 删除一条域数据
func (c *AppConnector) DeleteDomain(domainID int) error {
	err := c.execSQL(deleteDomainSQL, domainID)
	if err != nil {
		Error.Println(err)
		return err
	}
	return nil
}

//DeleteGroup 删除一条组数据
func (c *AppConnector) DeleteGroup(groupID int) error {
	err := c.execSQL(deleteGroupSQL, groupID)
	if err != nil {
		Error.Println(err)
		return err
	}
	return nil
}

//DeleteUser 删除一条用户数据
func (c *AppConnector) DeleteUser(userID int) error {
	err := c.execSQL(deleteUserSQL, userID)
	if err != nil {
		Error.Println(err)
		return err
	}
	return nil
}

//GetDomainIDByName 通过name查询Domain的ID
func (c *AppConnector) GetDomainIDByName(name string) (id int) {
	queryStr := fmt.Sprintf(selectDomainIDByNameSQL, name)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetDomainIDByName db query error:", err)
		return 0
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			Error.Println(err)
		}
	}
	return id
}

//GetRealmByID 通过ID获取domain的realm
func (c *AppConnector) GetRealmByID(id int) (realm string) {
	queryStr := fmt.Sprintf(selectRealmByIDSQL, id)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetDomainByID db query error:", err)
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&realm)
		if err != nil {
			Error.Println(err)
		}
	}
	return realm
}

//GetGroupIDByName 通过name查询group的ID
func (c *AppConnector) GetGroupIDByName(name string) (id int) {
	queryStr := fmt.Sprintf(selectGroupIDByNameSQL, name)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetGroupIDByName db query error:", err)
		return 0
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			Error.Println(err)
		}
	}
	return id
}

//GetGroupNameByID 通过ID获取group的name
func (c *AppConnector) GetGroupNameByID(id int) (name string) {
	queryStr := fmt.Sprintf(selectGroupNameByIDSQL, id)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetGroupNameByID db query error:", err)
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			Error.Println(err)
		}
	}
	return name
}

//GetOneGroupIDByDomainID 查找在域内的一个组ID
func (c *AppConnector) GetOneGroupIDByDomainID(id int) int {
	queryStr := fmt.Sprintf(selectOneGroupIDByDomainIDSQL, id)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetChildGroupIDbyParentID db query error:", err)
		return 0
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			Error.Println(err)
		}
	}
	return id
}

//GetChildGroupIDbyParentID 查找一个下级组的ID
func (c *AppConnector) GetChildGroupIDbyParentID(id int) int {

	queryStr := fmt.Sprintf(selectOneGroupIDByParentIDSQL, id)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetChildGroupIDbyParentID db query error:", err)
		return 0
	}
	defer rows.Close()
	childGroupID := 0
	for rows.Next() {
		err := rows.Scan(&childGroupID)
		if err != nil {
			Error.Println(err)
		}
	}
	return childGroupID
}

//GetOneUserByGroupID 查找一个组内用户ID
func (c *AppConnector) GetOneUserByGroupID(id int) int {
	queryStr := fmt.Sprintf(selectOneUserIDByGroupIDSQL, id)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("GetOneUserByGroupID db query error:", err)
		return 0
	}
	defer rows.Close()

	userID := 0
	for rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			Error.Println(err)
		}
	}
	return userID
}

//CheckUsernameInDomainExist 检查domain下username是否存在
func (c *AppConnector) CheckUsernameInDomainExist(domainID int, username string) bool {
	queryStr := fmt.Sprintf(selectUsernameFromDomainSQL, domainID, username)
	rows, err := c.dbConn.Query(queryStr)

	if err != nil {
		Error.Println("CheckUsernameInDomain db query error:", err)
		return true
	}
	defer rows.Close()
	var name string
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			Error.Println(err)
		}
	}
	if name == "" {
		return false
	}
	return true
}