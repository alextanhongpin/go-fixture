# go-fixture
An attempt to generate Postgres SQL for populating tables using yaml

## Tables

```sql
create table if not exists users(
	id int generated always as identity, 
	name text not null,
	age int not null,
	primary key(id)
);

create table if not exists accounts (
	id int generated always as identity,
	user_id int not null,
	type text not null,
	primary key(id),
	foreign key (user_id) references users(id)
);

create table if not exists authors  (
	id int generated always as identity,
	user_id int not null,
	penname text not null,
	primary key(id),
	foreign key (user_id) references users(id)
);

create table if not exists book_categories (
	id int generated always as identity,
	name text not null,
	primary key(id)
);


create table if not exists books (
	id int generated always as identity,
	author_id int not null,
	name text not null,
	book_category_id int not null,
	primary key (id),
	foreign key(author_id) references authors(id),
	foreign key(book_category_id) references book_categories(id)
);
```

## Input
```yaml
- table: users
  rows:
    - _id: smith
      name: smith
      age: 10
    - _id: john
      name: john
      age: 20
- table: accounts
  rows:
    - user_id: $.users.smith.id
      type: Facebook
    - user_id: $.users.smith.id
      type: Google
- table: books
  rows:
    - author_id: $.authors.smith.id
      name: Amazing Book
      book_category_id: $.book_categories.mystery.id
- table: authors
  rows:
    - _id: smith
      user_id: $.users.smith.id
      penname: smith
- table: book_categories
  rows:
    - _id: mystery
      name: mystery
```

Output:

```sql
WITH
users_smith AS (INSERT INTO users(age, name) VALUES (10, 'smith') RETURNING *),
users_john AS (INSERT INTO users(age, name) VALUES (20, 'john') RETURNING *),
authors_smith AS (INSERT INTO authors(penname, user_id) VALUES ('smith', (SELECT id FROM users_smith)) RETURNING *),
book_categories_mystery AS (INSERT INTO book_categories(name) VALUES ('mystery') RETURNING *),
books_0 AS (INSERT INTO books(author_id, book_category_id, name) VALUES ((SELECT id FROM authors_smith), (SELECT id FROM book_categories_mystery), 'Amazing Book') RETURNING *),
accounts_0 AS (INSERT INTO accounts(type, user_id) VALUES ('Facebook', (SELECT id FROM users_smith)) RETURNING *),
accounts_1 AS (INSERT INTO accounts(type, user_id) VALUES ('Google', (SELECT id FROM users_smith)) RETURNING *)
SELECT 1
```
