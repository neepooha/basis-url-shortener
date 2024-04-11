INSERT INTO urls (id, alias, url)
VALUES (1, 'habr', 'https://habr.com/')
ON CONFLICT DO NOTHING;