CREATE TABLE authentication (
	login VARCHAR(20) PRIMARY KEY,
	password VARCHAR(30) 
);

CREATE INDEX idx_auth ON authentication(login);
