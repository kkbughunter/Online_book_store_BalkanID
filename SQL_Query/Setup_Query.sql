
-- Create Database
CREATE USER myuser WITH PASSWORD 'mypassword';
CREATE DATABASE mydb;

-- Grant Permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.list TO myuser;

-- Create Table

CREATE TABLE admin(
    adminid VARCHAR(10),
    name VARCHAR(126),
    username VARCHAR(126),
    emailid VARCHAR(126),
    password VARCHAR(126),
    CONSTRAINT Admin_PK PRIMARY KEY (adminid),
    CONSTRAINT Admin_PK_syntax CHECK (adminid LIKE 'A%'),
    CONSTRAINT Admin_emailid_syntax CHECK (emailid LIKE'%@%.com')
);

CREATE TABLE books(
    ISBN INTEGER,
    title VARCHAR(126),
    author VARCHAR(126),
    genre VARCHAR(126),
    publisher VARCHAR(126),
    releasedata DATE,
    format VARCHAR(30) CHECK (format IN ('eBook', 'paperback', 'hardcover')),
    price NUMERIC(10, 2),
    stock INTEGER,
    CONSTRAINT book_PK PRIMARY KEY (bookid),
    CONSTRAINT book_Stock_check CHECK(stock >= 0)
);

CREATE TABLE user(
    userid VARCHAR(10),
    username VARCHAR(126),
    emailid VARCHAR(126),
    password VARCHAR(126),
    CONSTRAINT User_PK PRIMARY KEY (userid),
    CONSTRAINT User_PK_syntax CHECK (userid LIKE 'U%'),
    CONSTRAINT User_emailid_syntax CHECK (emailid LIKE'%@%.com')
);

CREATE TABLE wishlist(
    wishlistid SERIAL INTEGER,
    ISBN INTEGER,
    userid VARCHAR(10),
    timestmp TIMESTAMP,
    CONSTRAINT wishlist_PK PRIMARY KEY (wishlist),
    CONSTRAINT wishlist_FK_book FOREIGN KEY (ISBN) REFERENCES books(ISBN) ON DELETE CASCADE,
    CONSTRAINT wishlist_FK_user FOREIGN KEY (userid) REFERENCES user(userid) ON DELETE CASCADE
);

CREATE TABLE order(
    orderid SERIAL INTEGER,
    userid VARCHAR(10),
    orderdate DATE,
    totalamount INTEGER,
    status VARCHAR(30) CHECK (status IN ('processing', 'shipped', 'delivered')),
    CONSTRAINT order_FK_user FOREIGN KEY (userid) REFERENCES user(userid),
    CONSTRAINT order_check_amount CHECK( totalamount > 0)
);

CREATE TABLE orderdetail(
    orderdetailid SERIAL INTEGER,
    orderid INTEGER,
    ISBN VARCHAR(10),
    quantitiy INTEGER,
    subtotal INTEGER, -- (quantity * book price)
    CONSTRAINT OD_PK PRIMARY KEY (orderdetailid),
    CONSTRAINT OD_FK_book FOREIGN KEY (ISBN) REFERENCES books(ISBN) ON DELETE CASCADE,
    CONSTRAINT OD_FK_order FOREIGN KEY (orderid) REFERENCES order(orderid) ON DELETE CASCADE,
);

CREATE TABLE review(
    reviewid SERIAL INTEGER,
    ISBN VARCHAR(10),
    userid VARCHAR(10),
    rdate DATE,
    rating INTEGER CHECK(rating BETWEEN 1 AND 5),
    comment VARCHAR(512),
    CONSTRAINT review_PK PRIMARY KEY (reviewid),
    CONSTRAINT review_FK_book FOREIGN KEY (ISBN) REFERENCES books(ISBN) ON DELETE CASCADE,
    CONSTRAINT review_FK_user FOREIGN KEY (userid) REFERENCES user(userid) ON DELETE CASCADE
);