select
m.*,
a.*
from movies m
join movies_ages ma on ma.movie_id = m.id
join ages a on sa.age_id = a.id


select
m.*,
c.*
from movies m
join movies_categories mc on mc.movie_id = m.id
join categories c on mc.categorie_id = c.id


select
m.*,
g.*
from movies m
join movies_genres mg on mg.movie_id = m.id
join genres g on sg.genre_id = g.id

select
m.*,
e.*
from movies m
LEFT JOIN movies_allseries me ON me.movie_id = m.id
LEFT JOIN allseries e ON me.allserie_id = e.id



select
m.*,
s.*
from movies m
left join movies_seasons ms on ms.movie_id =m.id
left join season s on ms.season_id = s.id


select
s.*,
e.*
from season s
left JOIN seasons_allseries se ON se.season_id = s.id
left JOIN allseries e ON se.allserie_id = e.id
 