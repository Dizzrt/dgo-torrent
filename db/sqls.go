package db

const (
	_SQL_CREATE_TABLE_TASKS = `
		CREATE TABLE tasks (
			"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			"name" TEXT NOT NULL,
			"torrent" TEXT NOT NULL,
			"path" TEXT NOT NULL,
			"status" TEXT NOT NULL,
			"state" integer DEFAULT 0,
			"created_at" INTEGER DEFAULT (DATETIME(CURRENT_TIMESTAMP, 'localtime')),
			"updated_at" INTEGER DEFAULT (DATETIME(CURRENT_TIMESTAMP, 'localtime'))
	  	);
	  
	  	CREATE INDEX index_name
	  	ON tasks (
			"name" COLLATE BINARY ASC
	  	);
	  
	  	CREATE TRIGGER updated_trigger AFTER UPDATE 
	  	ON tasks
	  	FOR EACH ROW
	  	WHEN NEW.updated_at = OLD.updated_at
	  	BEGIN
			UPDATE tasks SET updated_at = (DATETIME(CURRENT_TIMESTAMP, 'localtime'))
			WHERE id = OLD.id;
	  	END;	  
	`
)
