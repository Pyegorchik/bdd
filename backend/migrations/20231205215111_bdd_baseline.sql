-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE public.auth_messages
(
    address    VARCHAR(42) NOT NULL,
    code       VARCHAR(255)   NOT NULL, 
    created_at BIGINT      NOT NULL,
    CONSTRAINT auth_messages_pkey PRIMARY KEY (address)
);


CREATE TABLE public.users
(
    id      BIGSERIAL  NOT NULL,
    role    INT       NOT NULL,
    address CHAR(42) NOT NULL,
    PRIMARY KEY (role, id)
);

CREATE TABLE public.jwtokens
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
            REFERENCES public.users (role, id)
);



ALTER TABLE public.auth_messages
    OWNER TO bdd;
ALTER TABLE public.users
    OWNER TO bdd;
ALTER TABLE public.jwtokens
    OWNER TO bdd;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE public.jwtokens;
DROP TABLE public.users;
DROP TABLE public.auth_messages;







