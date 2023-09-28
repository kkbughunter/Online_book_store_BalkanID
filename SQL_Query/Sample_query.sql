SELECT * FROM books;
INSERT INTO books VALUES (0001,'python','karthikeyan','BlakanID','Kumar','2023-05-03', 'hardcover', 1200, 4);

TRUNCATE TABLE books;
DELETE FROM books WHERE ISBN = 1;
DELETE FROM suser WHERE userid = 'UAdmin';