create table movies (
	id serial primary key,
 	title        text,
 	description  text,
	release_year int,
	director    text,
	rating      int default 0,
	is_watched  boolean,
	trailer_url  text,
	poster_url   text,
	viewsyt      int,
	duration     text,
	video_url    text,
	views_count  int;

);




create table allseries (
	id            serial primary key,
 	series       int,
 	title        text,
	trailer_url  text,
	poster_url   text
);

create table ages
	(
	  id serial primary key,
	  age text,
	  poster_url text
	);

create table genres
	(
	  id serial primary key,
	  title text,
	  poster_url text
	);
	
create table categories
	(
	  id serial primary key,
	  title text,
	  poster_url text
	);
		

create table movies_genres
(
    movie_id int references movies(id),
    genre_id int references genres(id)
);

	
create table movies_categories
(
    movie_id int references movies(id),
    categorie_id int references categories(id)
); 


create table movies_ages
(
    movie_id int references movie(id),
    age_id int references ages(id)
); 


create table movies_allseries
(
    movie_id int references movie(id),
    allserie_id int references allseries(id)
); 



ALTER TABLE movies_genres
ADD CONSTRAINT fk_movies_genres
FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE;

ALTER TABLE movies_categories
ADD CONSTRAINT fk_movies_categories
FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE;

ALTER TABLE movies_ages
ADD CONSTRAINT fk_movies_ages
FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE;

ALTER TABLE movies_allseries
ADD CONSTRAINT fk_movies_allsesries
FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE;
 ___________________________________________________________

 
create table season (
	id serial primary key,
	number int,
	tittle text,
)