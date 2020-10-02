CREATE TABLE Command (
    CommandID INTEGER PRIMARY KEY AUTOINCREMENT,
    Username  TEXT,
    RemoteIP  TEXT,
    Command   TEXT,
    Timestamp TEXT
);

CREATE TABLE Login (
    LoginID INTEGER PRIMARY KEY AUTOINCREMENT,
    Username      TEXT,
    Password      TEXT,
    RemoteIP      TEXT,
    RemoteVersion TEXT,
    Timestamp     TEXT
);
