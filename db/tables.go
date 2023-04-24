package db

// array of [2]string{SQL statement, context}
var setup = [][2]string{
	{sql_SETUP_XP, "creating the xp table"},
	{sql_TRIGGER_xp_func, "creating xp trigger func"},
	{sql_TRIGGER_XP, "creating xp trigger"},
	{sql_SETUP_WARN, "creating the warning table"},
	{sql_SETUP_GAMBLER, "creating the gambler table"},
	{sql_SETUP_TWITCH_COMMANDS, "creating the twitch command table"},
}

const sql_SETUP_XP = `CREATE TABLE IF NOT EXISTS xp (
	id TEXT PRIMARY KEY,
	xp INTEGER DEFAULT 0,
	lvl INTEGER DEFAULT 0,
	vc_time INTEGER DEFAULT 0,
	msg_num INTEGER DEFAULT 0,
	last_msg TIMESTAMPTZ DEFAULT NOW(),
	last_update TIMESTAMPTZ DEFAULT NOW()
)`

const sql_TRIGGER_xp_func = `CREATE OR REPLACE FUNCTION xp_trig_func() 
RETURNS TRIGGER 
LANGUAGE PLPGSQL
AS $$
BEGIN
	NEW.last_update = NOW();
	
	IF NEW.msg_num <> OLD.msg_num THEN
		 NEW.last_msg = NOW();
	END IF;

	RETURN NEW;
END;
$$`

const sql_TRIGGER_XP = `CREATE OR REPLACE TRIGGER trig_xp 
BEFORE UPDATE
	ON xp
	FOR EACH ROW
		EXECUTE PROCEDURE xp_trig_func()
`

// CREATE TRIGGER trigger_name 
//    {BEFORE | AFTER} { event }
//    ON table_name
//    [FOR [EACH] { ROW | STATEMENT }]
//        EXECUTE PROCEDURE trigger_function
const sql_SETUP_WARN = `CREATE TABLE IF NOT EXISTS warnings (
	id TEXT PRIMARY KEY,
	warned_user TEXT NOT NULL,
	severity INTEGER NOT NULL,
	reason TEXT NOT NULL
)`

const sql_SETUP_GAMBLER = `CREATE TABLE IF NOT EXISTS gambling (
	id TEXT,
	game INTEGER,
	lost INTEGER DEFAULT 0,
	won INTEGER DEFAULT 0,
	UNIQUE(id, game)
)`

const sql_SETUP_TWITCH_COMMANDS = `CREATE TABLE IF NOT EXISTS twitch_cmd (
	cmd TEXT,
	usr TEXT,
	resp TEXT,

	UNIQUE(cmd, usr)
)`
