-- Простые тестовые данные
INSERT INTO users (username, password, created_at) VALUES
('alice', 'password123', '2024-01-15 10:30:00'),
('bob', 'password456', '2024-01-15 11:15:00'),
('charlie', 'password789', '2024-01-15 12:00:00');

INSERT INTO chats (name, created_at) VALUES
('Общий чат', '2024-01-15 10:00:00'),
('Проект', '2024-01-15 11:00:00');

INSERT INTO chat_nembers (chat_id, user_id, joined_at) VALUES
((SELECT uuid FROM chats WHERE name = 'Общий чат' LIMIT 1), (SELECT id FROM users WHERE username = 'alice' LIMIT 1), NOW()),
((SELECT uuid FROM chats WHERE name = 'Общий чат' LIMIT 1), (SELECT id FROM users WHERE username = 'bob' LIMIT 1), NOW()),
((SELECT uuid FROM chats WHERE name = 'Проект' LIMIT 1), (SELECT id FROM users WHERE username = 'alice' LIMIT 1), NOW());

INSERT INTO message (chat_id, sender_id, content, sent_at) VALUES
((SELECT uuid FROM chats WHERE name = 'Общий чат' LIMIT 1), (SELECT id FROM users WHERE username = 'alice' LIMIT 1), 'Привет всем!', NOW()),
((SELECT uuid FROM chats WHERE name = 'Общий чат' LIMIT 1), (SELECT id FROM users WHERE username = 'bob' LIMIT 1), 'Привет!', NOW()),
((SELECT uuid FROM chats WHERE name = 'Проект' LIMIT 1), (SELECT id FROM users WHERE username = 'alice' LIMIT 1), 'Как дела с проектом?', NOW());

