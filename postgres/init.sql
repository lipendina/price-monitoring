CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS advert (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    link TEXT NOT NULL,
    name TEXT NOT NULL,
    price INT NOT NULL,
    is_removed BOOLEAN NOT NULL DEFAULT FALSE,
    last_check TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS confirmation (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    advert_id UUID REFERENCES advert(id) NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    is_confirm BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE UNIQUE INDEX advert_link_key ON advert(link) WHERE is_removed = FALSE;