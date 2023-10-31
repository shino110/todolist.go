-- Table for tasks
-- https://pkg.go.dev/time#example-Time.Format 
-- https://stackoverflow.com/questions/18598480/execute-formatted-time-in-a-slice-with-html-template
DROP TABLE IF EXISTS `tasks`;

CREATE TABLE `tasks` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `title` varchar(50) NOT NULL,
    `is_done` boolean NOT NULL DEFAULT b'0',
    `deadline` datetime,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `memo` varchar(256),
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;
