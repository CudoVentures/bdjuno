ALTER TABLE cw20token_info ADD type    TEXT NULL;
ALTER TABLE cw20token_info ADD creator TEXT NOT NULL;

ALTER TABLE cw20token_balance 
ALTER COLUMN balance TYPE TEXT;