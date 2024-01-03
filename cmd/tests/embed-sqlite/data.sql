-- Schema for the Library

-- Authors Table
CREATE TABLE authors (
    author_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    birth_date DATE
);

-- Books Table
CREATE TABLE books (
    book_id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    published_date DATE,
    author_id INTEGER,
    FOREIGN KEY(author_id) REFERENCES authors(author_id)
);

-- Categories Table
CREATE TABLE categories (
    category_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- Book-Categories Relationship Table
CREATE TABLE book_categories (
    book_id INTEGER,
    category_id INTEGER,
    PRIMARY KEY(book_id, category_id),
    FOREIGN KEY(book_id) REFERENCES books(book_id),
    FOREIGN KEY(category_id) REFERENCES categories(category_id)
);

-- Users Table
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    registration_date DATE
);

-- Borrowed Books Table
CREATE TABLE borrowed_books (
    user_id INTEGER,
    book_id INTEGER,
    borrow_date DATE,
    return_date DATE,
    PRIMARY KEY(user_id, book_id, borrow_date),
    FOREIGN KEY(user_id) REFERENCES users(user_id),
    FOREIGN KEY(book_id) REFERENCES books(book_id)
);

-- Inserting some sample data

-- Authors
INSERT INTO authors (name, birth_date) VALUES ('George Orwell', '1903-06-25');
INSERT INTO authors (name, birth_date) VALUES ('J.K. Rowling', '1965-07-31');

-- Books
INSERT INTO books (title, published_date, author_id) VALUES ('1984', '1949-06-08', 1);
INSERT INTO books (title, published_date, author_id) VALUES ('Animal Farm', '1945-08-17', 1);
INSERT INTO books (title, published_date, author_id) VALUES ('Harry Potter and the Philosopher''s Stone', '1997-06-26', 2);

-- Categories
INSERT INTO categories (name) VALUES ('Dystopia');
INSERT INTO categories (name) VALUES ('Fantasy');
INSERT INTO categories (name) VALUES ('Political');

-- Book-Categories
INSERT INTO book_categories (book_id, category_id) VALUES (1, 1);
INSERT INTO book_categories (book_id, category_id) VALUES (1, 3);
INSERT INTO book_categories (book_id, category_id) VALUES (2, 3);
INSERT INTO book_categories (book_id, category_id) VALUES (3, 2);

-- Users
INSERT INTO users (name, registration_date) VALUES ('Alice Smith', '2022-01-15');
INSERT INTO users (name, registration_date) VALUES ('Bob Johnson', '2020-12-20');

-- Borrowed Books
INSERT INTO borrowed_books (user_id, book_id, borrow_date, return_date) VALUES (1, 1, '2022-01-20', '2022-02-15');
INSERT INTO borrowed_books (user_id, book_id, borrow_date, return_date) VALUES (1, 3, '2022-03-01', NULL);

