CREATE TABLE exchange_rates (
	id SERIAL PRIMARY KEY,
	date DATE NOT NULL,
	base_currency VARCHAR(10) NOT NULL DEFAULT 'RUB',
	target_currency VARCHAR(10) NOT NULL,
	currency_rate REAL NOT NULL
);

CREATE INDEX idx_exchange_rates_date_base_currency ON exchange_rates(date, base_currency);
