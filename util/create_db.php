<?php
unlink(".cache/caffo.db");
$sql = new SQLite3(".cache/caffo.db");
$sql->query("
CREATE TABLE `plays` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `file` TEXT NOT NULL,
    `ip` TEXT NOT NULL,
    `time` INTEGER NOT NULL
);

CREATE TABLE `downloads` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `file` TEXT NOT NULL,
    `ip` TEXT NOT NULL,
    `time` INTEGER NOT NULL
);
");

