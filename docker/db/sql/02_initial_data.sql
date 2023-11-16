INSERT INTO `tasks` (`title`) VALUES ("sample-task-01");
INSERT INTO `tasks` (`title`) VALUES ("sample-task-02");
INSERT INTO `tasks` (`title`, `deadline`) VALUES ("sample-task-03", "23-10-6 19:30:00");
INSERT INTO `tasks` (`title`, `memo`) VALUES ("sample-task-04", "memomemo");
INSERT INTO `tasks` (`title`, `is_done`) VALUES ("sample-task-05", true);

INSERT INTO `ownership` (`user_id`, `task_id`) VALUES (1,1);
INSERT INTO `ownership` (`user_id`, `task_id`) VALUES (1,2);
INSERT INTO `ownership` (`user_id`, `task_id`) VALUES (1,3);
INSERT INTO `ownership` (`user_id`, `task_id`) VALUES (1,4);