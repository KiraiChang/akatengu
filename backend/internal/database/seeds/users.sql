INSERT INTO users (username, password, status)
VALUES
    ("admin","$argon2id$v=19$m=65536,t=3,p=4$dAwHffvcmwhMyBdu17BV2A$zKcGOexREzx6GetSdqq6OwXmOHWP5cY1+MRyTV9iq/c","ACTIVE")
    ON CONFLICT (username) DO NOTHING;