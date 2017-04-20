<?php
/**
 * usage: php -f util/create_db.php > .cache/caffo.db
 *
 * Easiest way to create an sqlite db imo
 * Feel free to copypaste the sql below instead
 */
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

