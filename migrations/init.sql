CREATE TABLE users (
    id SERIAL4 NOT NULL PRIMARY KEY,
    username VARCHAR(255) DEFAULT NULL,
    password VARCHAR(255) DEFAULT NULL,
    created_at VARCHAR(200) DEFAULT NULL,
    public_key TEXT DEFAULT NULL
);

CREATE TABLE chats (
    uuid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE chat_numbers (
    chat_id UUID NOT NULL,
    user_id INT4 NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE message (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    chat_id UUID NOT NULL,
    sender_id INT4 NOT NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE message_keys (
    id SERIAL4 NOT NULL PRIMARY KEY,
    message_id INT4 NOT NULL,
    user_id INT4 NOT NULL,
    encrypted_key TEXT DEFAULT NULL
);

CREATE TABLE message_status (
    message_id INT8 NOT NULL,
    user_id INT4 NOT NULL,
    status VARCHAR(255) DEFAULT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (message_id, user_id)
);

CREATE TABLE ai_chats (
    id SERIAL4 NOT NULL PRIMARY KEY,
    user_id INT4 NOT NULL,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE ai_messages (
    id SERIAL4 NOT NULL PRIMARY KEY,
    chat_id INT4 NOT NULL,
    content TEXT NOT NULL,
    sender_type VARCHAR(4) NOT NULL,
    sent_at TIMESTAMP DEFAULT now()
);