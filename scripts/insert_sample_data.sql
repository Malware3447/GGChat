-- Тестовые данные для демонстрации визуализации БД

-- Очищаем существующие данные
TRUNCATE TABLE message_status CASCADE;
TRUNCATE TABLE message CASCADE;
TRUNCATE TABLE chat_nembers CASCADE;
TRUNCATE TABLE chats CASCADE;
TRUNCATE TABLE users CASCADE;

-- Добавляем пользователей
INSERT INTO users (username, password, created_at) VALUES
('alice', 'password123', '2024-01-15 10:30:00'),
('bob', 'password456', '2024-01-15 11:15:00'),
('charlie', 'password789', '2024-01-15 12:00:00'),
('diana', 'password000', '2024-01-16 09:45:00');

-- Добавляем чаты
INSERT INTO chats (name, created_at) VALUES
('Общий чат команды', '2024-01-15 10:00:00'),
('Разработка проекта', '2024-01-15 11:00:00'),
('Личные сообщения', '2024-01-16 08:30:00');

-- Добавляем участников чатов
INSERT INTO chat_nembers (chat_id, user_id, joined_at) VALUES
((SELECT uuid FROM chats WHERE name = 'Общий чат команды'), (SELECT id FROM users WHERE username = 'alice'), '2024-01-15 10:00:00'),
((SELECT uuid FROM chats WHERE name = 'Общий чат команды'), (SELECT id FROM users WHERE username = 'bob'), '2024-01-15 10:05:00'),
((SELECT uuid FROM chats WHERE name = 'Общий чат команды'), (SELECT id FROM users WHERE username = 'charlie'), '2024-01-15 10:10:00'),
((SELECT uuid FROM chats WHERE name = 'Разработка проекта'), (SELECT id FROM users WHERE username = 'alice'), '2024-01-15 11:00:00'),
((SELECT uuid FROM chats WHERE name = 'Разработка проекта'), (SELECT id FROM users WHERE username = 'diana'), '2024-01-16 09:45:00'),
((SELECT uuid FROM chats WHERE name = 'Личные сообщения'), (SELECT id FROM users WHERE username = 'alice'), '2024-01-16 08:30:00'),
((SELECT uuid FROM chats WHERE name = 'Личные сообщения'), (SELECT id FROM users WHERE username = 'bob'), '2024-01-16 08:30:00');

-- Добавляем сообщения
INSERT INTO message (chat_id, sender_id, content, sent_at) VALUES
((SELECT uuid FROM chats WHERE name = 'Общий чат команды'), (SELECT id FROM users WHERE username = 'alice'), 'Привет всем! Как дела с проектом?', '2024-01-15 10:15:00'),
((SELECT uuid FROM chats WHERE name = 'Общий чат команды'), (SELECT id FROM users WHERE username = 'bob'), 'Привет! Всё идёт по плану, спасибо!', '2024-01-15 10:16:00'),
((SELECT uuid FROM chats WHERE name = 'Общий чат команды'), (SELECT id FROM users WHERE username = 'charlie'), 'Да, прогресс есть. Скоро покажем демо', '2024-01-15 10:17:00'),
((SELECT uuid FROM chats WHERE name = 'Разработка проекта'), (SELECT id FROM users WHERE username = 'alice'), 'Diana, как продвигается работа над API?', '2024-01-16 09:50:00'),
((SELECT uuid FROM chats WHERE name = 'Разработка проекта'), (SELECT id FROM users WHERE username = 'diana'), 'API готов на 80%, завтра доделаю остальное', '2024-01-16 09:52:00'),
((SELECT uuid FROM chats WHERE name = 'Личные сообщения'), (SELECT id FROM users WHERE username = 'alice'), 'Bob, можешь помочь с кодом?', '2024-01-16 10:00:00'),
((SELECT uuid FROM chats WHERE name = 'Личные сообщения'), (SELECT id FROM users WHERE username = 'bob'), 'Конечно! Что именно нужно?', '2024-01-16 10:02:00');

-- Добавляем статусы сообщений
INSERT INTO message_status (message_id, user_id, status, updated_at) VALUES
(1, (SELECT id FROM users WHERE username = 'bob'), 'read', '2024-01-15 10:16:00'),
(1, (SELECT id FROM users WHERE username = 'charlie'), 'read', '2024-01-15 10:17:00'),
(2, (SELECT id FROM users WHERE username = 'alice'), 'read', '2024-01-15 10:17:00'),
(2, (SELECT id FROM users WHERE username = 'charlie'), 'read', '2024-01-15 10:18:00'),
(3, (SELECT id FROM users WHERE username = 'alice'), 'read', '2024-01-15 10:18:00'),
(3, (SELECT id FROM users WHERE username = 'bob'), 'read', '2024-01-15 10:18:00'),
(4, (SELECT id FROM users WHERE username = 'diana'), 'read', '2024-01-16 09:52:00'),
(5, (SELECT id FROM users WHERE username = 'alice'), 'read', '2024-01-16 09:53:00'),
(6, (SELECT id FROM users WHERE username = 'bob'), 'read', '2024-01-16 10:02:00'),
(7, (SELECT id FROM users WHERE username = 'alice'), 'read', '2024-01-16 10:03:00);
