INSERT INTO urls (alias, url)
VALUES ('habr', 'https://habr.com/')
ON CONFLICT DO NOTHING;