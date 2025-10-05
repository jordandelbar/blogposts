CREATE SCHEMA IF NOT EXISTS content;

CREATE TABLE content.articles (
    id serial PRIMARY KEY,
    title text NOT NULL,
    slug text UNIQUE NOT NULL,
    content text NOT NULL,
    created_at timestamp(0) with time zone DEFAULT now(),
    updated_at timestamp(0) with time zone DEFAULT now(),
    published_at timestamp(0) with time zone,
    deleted_at timestamp(0) with time zone,
    is_published boolean DEFAULT false,
    is_deleted boolean DEFAULT false
);
