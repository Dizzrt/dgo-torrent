CREATE TABLE works (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL,
  "torrent" TEXT NOT NULL,
  "path" TEXT NOT NULL,
  "stat" integer DEFAULT 0,
  "created_at" INTEGER DEFAULT (DATETIME(CURRENT_TIMESTAMP, 'localtime')),
  "updated_at" INTEGER DEFAULT (DATETIME(CURRENT_TIMESTAMP, 'localtime'))
);

CREATE INDEX index_name
ON works (
	"name" COLLATE BINARY ASC
);

CREATE TRIGGER updated_trigger AFTER UPDATE 
ON works
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
	UPDATE works SET updated_at = (DATETIME(CURRENT_TIMESTAMP, 'localtime'))
	WHERE id = OLD.id;
END;
