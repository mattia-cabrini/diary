/* SPDX-License-Identifier: MIT */

PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;

CREATE TABLE entries (
    id INTEGER primary key AUTOINCREMENT,
    init INTEGER,
    fin INTEGER,
    inserted INTEGER,
    note text,
    deleted INTEGER
);

CREATE TABLE attachments (
    id INTEGER primary key AUTOINCREMENT,
    name TEXT,
    inserted INTEGER,
    content BLOB,
    entry_id INTEGER,
    FOREIGN KEY(entry_id) REFERENCES entries(id)
);

/* To manually log inconsistencies due to tests or errors */
CREATE TABLE anomalies (
    inserted INTEGER,
    note TEXT
);


COMMIT;
