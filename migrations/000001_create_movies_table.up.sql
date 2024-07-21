CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    year integer NOT NULL,
    runtime integer NOT NULL,
    genres text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);

INSERT INTO movies (title, year, runtime, genres)
VALUES
    ('Inception', 2010, 148, ARRAY['Action', 'Sci-Fi']),
    ('The Matrix', 1999, 136, ARRAY['Action', 'Sci-Fi']),
    ('Interstellar', 2014, 169, ARRAY['Adventure', 'Drama', 'Sci-Fi']),
    ('The Godfather', 1972, 175, ARRAY['Crime', 'Drama']),
    ('Pulp Fiction', 1994, 154, ARRAY['Crime', 'Drama']);