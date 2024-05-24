-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE public.auth_messages_chain
(
    address    VARCHAR(42) NOT NULL,
    code       VARCHAR(255)   NOT NULL, 
    created_at BIGINT      NOT NULL,
    CONSTRAINT auth_messages_pkey PRIMARY KEY (address)
);


CREATE TABLE public.users_chain
(
    id      BIGSERIAL  NOT NULL,
    role    INT       NOT NULL,
    address CHAR(42) NOT NULL,
    PRIMARY KEY (role, id)
);

CREATE TABLE public.jwtokens_chain
(
    id         BIGINT       NOT NULL,
    purpose    INT          NOT NULL,
    role       INT          NOT NULL,
    number     INT          NOT NULL,
    expires_at TIMESTAMP    NOT NULL,
    secret     VARCHAR(255) NOT NULL,
    PRIMARY KEY (role, id, number, purpose),
    CONSTRAINT fk_users
        FOREIGN KEY (role, id)
            REFERENCES public.users_chain (role, id)
);

-- Создание таблицы пользователей
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
);

-- Создание таблицы диалогов
CREATE TABLE dialogs (
    id BIGSERIAL PRIMARY KEY,
);

-- Создание таблицы участников диалогов
CREATE TABLE dialog_participants (
    dialog_id INT NOT NULL,
    user_id INT NOT NULL,
    PRIMARY KEY (dialog_id, user_id),
    FOREIGN KEY (dialog_id) REFERENCES dialogs(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Создание таблицы сообщений
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    dialog_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (dialog_id) REFERENCES dialogs(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Индекс для быстрого поиска сообщений по диалогу
CREATE INDEX idx_messages_dialog_id ON messages(dialog_id);

-- Индекс для быстрого поиска участников диалога
CREATE INDEX idx_dialog_participants_user_id ON dialog_participants(user_id);



ALTER TABLE public.auth_messages_chain
    OWNER TO bdd;
ALTER TABLE public.users_chain
    OWNER TO bdd;
ALTER TABLE public.jwtokens_chain
    OWNER TO bdd;
ALTER TABLE public.users
    OWNER TO bdd;
ALTER TABLE public.dialogs 
    OWNER TO bdd;
ALTER TABLE public.dialog_participants 
    OWNER TO bdd;
ALTER TABLE public.messages 
    OWNER TO bdd;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE public.jwtokens_chain;
DROP TABLE public.users_chain;
DROP TABLE public.auth_messages_chain;

DROP TABLE public.users;
DROP INDEX IF EXISTS idx_messages_dialog_id;
DROP INDEX IF EXISTS idx_dialog_participants_user_id;
DROP TABLE IF EXISTS public.messages;
DROP TABLE IF EXISTS public.dialog_participants;
DROP TABLE IF EXISTS public.dialogs;





