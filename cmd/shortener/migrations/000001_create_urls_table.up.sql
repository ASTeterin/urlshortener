CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL,
    original_url VARCHAR(255) NOT NULL
);