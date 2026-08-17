package migrations

var allMigrations = []Migration{
	{
		ID:       "0001_initial_schema",
		SQLite:   initialSchemaSQLite,
		Postgres: initialSchemaPostgres,
	},
}

var initialSchemaSQLite = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id integer PRIMARY KEY AUTOINCREMENT,
		first_name text NOT NULL,
		second_name text NOT NULL,
		nickname text NOT NULL,
		password_hash text NOT NULL,
		role text DEFAULT 'author',
		created_at datetime,
		registration_ip text,
		last_login_ip text,
		last_login_at datetime,
		is_blocked numeric DEFAULT false,
		warn_count integer DEFAULT 0
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname)`,
	`CREATE TABLE IF NOT EXISTS warnings (
		id integer PRIMARY KEY AUTOINCREMENT,
		user_id integer,
		reason text,
		created_at datetime,
		expires_at datetime,
		CONSTRAINT fk_users_warnings FOREIGN KEY (user_id) REFERENCES users(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_warnings_user_id ON warnings(user_id)`,
	`CREATE TABLE IF NOT EXISTS game_infos (
		id integer PRIMARY KEY AUTOINCREMENT,
		title text,
		icon_url text
	)`,
	`CREATE TABLE IF NOT EXISTS game_cards (
		id integer PRIMARY KEY AUTOINCREMENT,
		title text,
		icon_url text,
		game_id integer
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_cards_game_id ON game_cards(game_id)`,
	`CREATE TABLE IF NOT EXISTS translate_cards (
		id integer PRIMARY KEY AUTOINCREMENT,
		author_name text,
		author_id integer,
		source text,
		version real,
		percent_ready real,
		url_to_download text,
		archive_hash text,
		file_size real,
		status text DEFAULT 'pending',
		scan_details text DEFAULT '',
		game_files text,
		game_info_id integer,
		created_at datetime,
		CONSTRAINT fk_game_infos_translate_cards FOREIGN KEY (game_info_id) REFERENCES game_infos(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_archive_hash ON translate_cards(archive_hash)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_status ON translate_cards(status)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_game_info_id ON translate_cards(game_info_id)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_author_id ON translate_cards(author_id)`,
	`CREATE TABLE IF NOT EXISTS notifications (
		id integer PRIMARY KEY AUTOINCREMENT,
		user_id integer,
		title text NOT NULL,
		message text NOT NULL,
		is_read numeric DEFAULT false,
		created_at datetime
	)`,
	`CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)`,
	`CREATE TABLE IF NOT EXISTS chat_messages (
		id integer PRIMARY KEY AUTOINCREMENT,
		from_user_id integer,
		to_user_id integer,
		encrypted_text blob,
		is_read numeric DEFAULT false,
		created_at datetime
	)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_messages_from_user_id ON chat_messages(from_user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_messages_to_user_id ON chat_messages(to_user_id)`,
	`CREATE TABLE IF NOT EXISTS events (
		id integer PRIMARY KEY AUTOINCREMENT,
		actor_id integer,
		actor_role text,
		action text,
		target_type text,
		target_id integer,
		details text,
		ip text,
		user_agent text,
		created_at datetime
	)`,
	`CREATE INDEX IF NOT EXISTS idx_events_actor_id ON events(actor_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_action ON events(action)`,
	`CREATE INDEX IF NOT EXISTS idx_events_target_type ON events(target_type)`,
	`CREATE INDEX IF NOT EXISTS idx_events_target_id ON events(target_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at)`,
}

var initialSchemaPostgres = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id bigserial PRIMARY KEY,
		first_name text NOT NULL,
		second_name text NOT NULL,
		nickname text NOT NULL,
		password_hash text NOT NULL,
		role text DEFAULT 'author',
		created_at timestamptz,
		registration_ip text,
		last_login_ip text,
		last_login_at timestamptz,
		is_blocked boolean DEFAULT false,
		warn_count bigint DEFAULT 0
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname)`,
	`CREATE TABLE IF NOT EXISTS warnings (
		id bigserial PRIMARY KEY,
		user_id bigint,
		reason text,
		created_at timestamptz,
		expires_at timestamptz,
		CONSTRAINT fk_users_warnings FOREIGN KEY (user_id) REFERENCES users(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_warnings_user_id ON warnings(user_id)`,
	`CREATE TABLE IF NOT EXISTS game_infos (
		id bigserial PRIMARY KEY,
		title text,
		icon_url text
	)`,
	`CREATE TABLE IF NOT EXISTS game_cards (
		id bigserial PRIMARY KEY,
		title text,
		icon_url text,
		game_id bigint
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_cards_game_id ON game_cards(game_id)`,
	`CREATE TABLE IF NOT EXISTS translate_cards (
		id bigserial PRIMARY KEY,
		author_name text,
		author_id bigint,
		source text,
		version double precision,
		percent_ready double precision,
		url_to_download text,
		archive_hash text,
		file_size double precision,
		status text DEFAULT 'pending',
		scan_details text DEFAULT '',
		game_files jsonb,
		game_info_id bigint,
		created_at timestamptz,
		CONSTRAINT fk_game_infos_translate_cards FOREIGN KEY (game_info_id) REFERENCES game_infos(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_archive_hash ON translate_cards(archive_hash)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_status ON translate_cards(status)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_game_info_id ON translate_cards(game_info_id)`,
	`CREATE INDEX IF NOT EXISTS idx_translate_cards_author_id ON translate_cards(author_id)`,
	`CREATE TABLE IF NOT EXISTS notifications (
		id bigserial PRIMARY KEY,
		user_id bigint,
		title text NOT NULL,
		message text NOT NULL,
		is_read boolean DEFAULT false,
		created_at timestamptz
	)`,
	`CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)`,
	`CREATE TABLE IF NOT EXISTS chat_messages (
		id bigserial PRIMARY KEY,
		from_user_id bigint,
		to_user_id bigint,
		encrypted_text bytea,
		is_read boolean DEFAULT false,
		created_at timestamptz
	)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_messages_from_user_id ON chat_messages(from_user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_messages_to_user_id ON chat_messages(to_user_id)`,
	`CREATE TABLE IF NOT EXISTS events (
		id bigserial PRIMARY KEY,
		actor_id bigint,
		actor_role text,
		action text,
		target_type text,
		target_id bigint,
		details text,
		ip text,
		user_agent text,
		created_at timestamptz
	)`,
	`CREATE INDEX IF NOT EXISTS idx_events_actor_id ON events(actor_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_action ON events(action)`,
	`CREATE INDEX IF NOT EXISTS idx_events_target_type ON events(target_type)`,
	`CREATE INDEX IF NOT EXISTS idx_events_target_id ON events(target_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at)`,
}
