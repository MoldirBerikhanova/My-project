ALTER TABLE allseries
ADD COLUMN duration INT;

ALTER TABLE allseries
ADD COLUMN poster_url text;

ALTER TABLE allseries
drop COLUMN description, 
drop column release_year, 
drop column director, 
drop column rating;