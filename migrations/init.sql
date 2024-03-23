CREATE EXTENSION IF NOT EXISTS pg_trgm;

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_roles') THEN
            CREATE TYPE user_roles AS ENUM ('admin', 'jobseeker', 'employer');
        END IF;
END$$;

CREATE TABLE IF NOT EXISTS users(
    full_name char(100) not null,
    passhash text not null,
    phone char(50) not null unique,
    email char(100) not null unique,
    user_role user_roles not null,
    id bigserial primary key,
    deleted bool not null default false,
    created_at timestamp not null default now()
);

CREATE INDEX IF NOT EXISTS idx_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_phone ON users (phone);

CREATE TABLE IF NOT EXISTS forget_password_info(
    id     bigserial PRIMARY KEY,
    link   text NOT NULL UNIQUE,
    foreign key (user_id) references users(id),
    user_id bigint NOT NULL,
    expiration timestamp NOT NULL,
    created_at timestamp not null default now()
);

CREATE INDEX IF NOT EXISTS idx_link ON forget_password_info(link);