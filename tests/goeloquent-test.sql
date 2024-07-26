/*
 Navicat Premium Data Transfer

 Source Server         : mamplocal
 Source Server Type    : MySQL
 Source Server Version : 50739
 Source Schema         : goeloquent

 Target Server Type    : MySQL
 Target Server Version : 50739
 File Encoding         : 65001

 Date: 26/07/2024 16:22:05
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for comments
-- ----------------------------
DROP TABLE IF EXISTS `comments`;
CREATE TABLE `comments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content` varchar(255) NOT NULL,
  `likes` int(11) NOT NULL,
  `commentable_id` int(11) NOT NULL,
  `commentable_type` varchar(255) NOT NULL,
  `parent_id` int(11) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of comments
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for images
-- ----------------------------
DROP TABLE IF EXISTS `images`;
CREATE TABLE `images` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `path` varchar(255) NOT NULL,
  `size` int(11) NOT NULL,
  `driver` varchar(255) NOT NULL,
  `imageable_id` int(11) NOT NULL,
  `imageable_type` varchar(255) NOT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of images
-- ----------------------------
BEGIN;
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (1, 'a1/b1/c1.png', 1024, 's3', 3, 'user', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (2, 'a2/b2/c2.png', 2048, 'gcp', 1, 'post', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (3, 'a3/b3/c3.png', 3072, 's3', 1, 'user', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (4, 'a4/b4/c4.png', 4096, 's3', 2, 'user', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (5, 'a5/b5/c5.png', 5120, 'gcp', 2, 'post', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (6, 'a6/b6/c6.png', 6144, 's3', 3, 'post', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (7, 'a7/b7/c7.png', 7168, 's3', 3, 'user', NULL, NULL, NULL, NULL);
INSERT INTO `images` (`id`, `path`, `size`, `driver`, `imageable_id`, `imageable_type`, `remark`, `created_at`, `updated_at`, `deleted_at`) VALUES (8, 'a8/b8/c8.png', 8192, 's3', 3, 'post', NULL, NULL, NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for logs
-- ----------------------------
DROP TABLE IF EXISTS `logs`;
CREATE TABLE `logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `logable_id` int(11) DEFAULT NULL,
  `logable_type` varchar(255) DEFAULT NULL,
  `operator_id` int(11) DEFAULT NULL,
  `operator_type` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `meta` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of logs
-- ----------------------------
BEGIN;
INSERT INTO `logs` (`id`, `logable_id`, `logable_type`, `operator_id`, `operator_type`, `created_at`, `remark`, `meta`) VALUES (1, 1, 'user', 1, NULL, NULL, 'user created', '{\"Verified\":false,\"Age\":0,\"Address\":\"\",\"Links\":null}');
INSERT INTO `logs` (`id`, `logable_id`, `logable_type`, `operator_id`, `operator_type`, `created_at`, `remark`, `meta`) VALUES (2, 1, 'user', 1, NULL, NULL, 'user saved', '{\"Verified\":false,\"Age\":0,\"Address\":\"\",\"Links\":null}');
INSERT INTO `logs` (`id`, `logable_id`, `logable_type`, `operator_id`, `operator_type`, `created_at`, `remark`, `meta`) VALUES (3, 1, 'user', 1, NULL, NULL, 'user updated', '{\"Verified\":false,\"Age\":0,\"Address\":\"\",\"Links\":null}');
INSERT INTO `logs` (`id`, `logable_id`, `logable_type`, `operator_id`, `operator_type`, `created_at`, `remark`, `meta`) VALUES (4, 1, 'user', 1, NULL, NULL, 'user saved', '{\"Verified\":false,\"Age\":0,\"Address\":\"\",\"Links\":null}');
INSERT INTO `logs` (`id`, `logable_id`, `logable_type`, `operator_id`, `operator_type`, `created_at`, `remark`, `meta`) VALUES (5, 1, 'user', 1, NULL, NULL, 'user retrieved', '');
INSERT INTO `logs` (`id`, `logable_id`, `logable_type`, `operator_id`, `operator_type`, `created_at`, `remark`, `meta`) VALUES (6, 1, 'user', 1, NULL, NULL, 'user deleted', '');
COMMIT;

-- ----------------------------
-- Table structure for notifications
-- ----------------------------
DROP TABLE IF EXISTS `notifications`;
CREATE TABLE `notifications` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `content` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `read_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of notifications
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Table structure for phones
-- ----------------------------
DROP TABLE IF EXISTS `phones`;
CREATE TABLE `phones` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `country` varchar(255) NOT NULL,
  `tel` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of phones
-- ----------------------------
BEGIN;
INSERT INTO `phones` (`id`, `user_id`, `country`, `tel`) VALUES (1, 5, '+1', '97812345678');
INSERT INTO `phones` (`id`, `user_id`, `country`, `tel`) VALUES (2, 4, '+1', '9712345655');
INSERT INTO `phones` (`id`, `user_id`, `country`, `tel`) VALUES (3, 1, '+2', '123456789');
INSERT INTO `phones` (`id`, `user_id`, `country`, `tel`) VALUES (4, 3, '+3', '1563103');
INSERT INTO `phones` (`id`, `user_id`, `country`, `tel`) VALUES (5, 2, '+1', '1231241234');
COMMIT;

-- ----------------------------
-- Table structure for posts
-- ----------------------------
DROP TABLE IF EXISTS `posts`;
CREATE TABLE `posts` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `title` varchar(255) DEFAULT NULL,
  `meta` json DEFAULT NULL,
  `tags` varchar(255) DEFAULT NULL,
  `status` tinyint(4) DEFAULT NULL,
  `created_time` timestamp NULL DEFAULT NULL,
  `updated_time` timestamp NULL DEFAULT NULL,
  `deleted_time` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of posts
-- ----------------------------
BEGIN;
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (1, 5, 'user5', NULL, '[golang,mysql,orm]', 1, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (2, 5, 'user5', NULL, '[eloquent,mysql,golang]', 2, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (3, 4, 'user4', NULL, '[relation,hasMany,golang]', 1, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (4, 2, 'user2', NULL, '[doc,readme]', 1, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (5, 1, 'user1', NULL, '[model,relation]', 2, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (6, 1, 'user1', NULL, '[test,golang]', 1, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (7, 2, 'user2', NULL, '[constraints,laravel]', 2, NULL, NULL, NULL);
INSERT INTO `posts` (`id`, `user_id`, `title`, `meta`, `tags`, `status`, `created_time`, `updated_time`, `deleted_time`) VALUES (8, 4, 'user4', NULL, '[bindings,wheres]', 0, NULL, NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for role_users
-- ----------------------------
DROP TABLE IF EXISTS `role_users`;
CREATE TABLE `role_users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `role_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `status` tinyint(4) DEFAULT NULL,
  `meta` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `expired_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of role_users
-- ----------------------------
BEGIN;
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (1, 1, 2, 1, '{\"remark\":\"asd\",\"granted_by\":0}', NULL, NULL, NULL, NULL);
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (2, 1, 1, 1, '{\"remark\":\"asd\",\"granted_by\":0}', NULL, NULL, NULL, NULL);
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (3, 1, 3, 0, '{\"remark\":\"asd\",\"granted_by\":0,\"revoked_by\":1}', NULL, NULL, NULL, NULL);
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (4, 2, 4, 1, '{\"remark\":\"asd\",\"granted_by\":0}', NULL, NULL, NULL, NULL);
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (5, 2, 5, 1, '{\"remark\":\"asd\",\"granted_by\":0}', NULL, NULL, NULL, NULL);
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (6, 3, 2, 1, '{\"remark\":\"asd\",\"granted_by\":0}', NULL, NULL, NULL, NULL);
INSERT INTO `role_users` (`id`, `role_id`, `user_id`, `status`, `meta`, `created_at`, `updated_at`, `expired_at`, `deleted_at`) VALUES (7, 3, 3, 0, '{\"remark\":\"asd\",\"granted_by\":0,\"revoked_by\":1}', NULL, NULL, NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for roles
-- ----------------------------
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `permissions` varchar(255) NOT NULL COMMENT 'comma seperated, used for scanner/valuer test',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of roles
-- ----------------------------
BEGIN;
INSERT INTO `roles` (`id`, `name`, `permissions`) VALUES (1, 'admin', 'create,update,delete,query');
INSERT INTO `roles` (`id`, `name`, `permissions`) VALUES (2, 'clerk', 'create,update');
INSERT INTO `roles` (`id`, `name`, `permissions`) VALUES (3, 'member', 'query');
COMMIT;

-- ----------------------------
-- Table structure for tagables
-- ----------------------------
DROP TABLE IF EXISTS `tagables`;
CREATE TABLE `tagables` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tag_id` int(11) NOT NULL,
  `tagable_id` int(11) NOT NULL,
  `tagable_type` varchar(255) DEFAULT NULL,
  `show_in_list` int(11) DEFAULT NULL,
  `tag_url` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of tagables
-- ----------------------------
BEGIN;
INSERT INTO `tagables` (`id`, `tag_id`, `tagable_id`, `tagable_type`, `show_in_list`, `tag_url`, `created_at`, `updated_at`) VALUES (1, 1, 1, 'post', 1, 't1url', NULL, NULL);
INSERT INTO `tagables` (`id`, `tag_id`, `tagable_id`, `tagable_type`, `show_in_list`, `tag_url`, `created_at`, `updated_at`) VALUES (2, 2, 2, 'post', 1, 't2url', NULL, NULL);
INSERT INTO `tagables` (`id`, `tag_id`, `tagable_id`, `tagable_type`, `show_in_list`, `tag_url`, `created_at`, `updated_at`) VALUES (3, 1, 3, 'post', 0, 't1url', NULL, NULL);
INSERT INTO `tagables` (`id`, `tag_id`, `tagable_id`, `tagable_type`, `show_in_list`, `tag_url`, `created_at`, `updated_at`) VALUES (4, 2, 8, 'post', 1, 't2url', NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for tags
-- ----------------------------
DROP TABLE IF EXISTS `tags`;
CREATE TABLE `tags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `related` int(11) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of tags
-- ----------------------------
BEGIN;
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (1, 'doc', 1, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (2, 'eloquent', 1, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (3, 'mysql', 1, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (4, 'golang', 3, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (5, 'orm', 1, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (6, 'relation', 2, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (7, 'hasmany', 1, NULL, NULL, NULL);
INSERT INTO `tags` (`id`, `name`, `related`, `created_at`, `updated_at`, `deleted_at`) VALUES (8, 'hasone', 0, NULL, NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for user_models
-- ----------------------------
DROP TABLE IF EXISTS `user_models`;
CREATE TABLE `user_models` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `no` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `status` tinyint(4) DEFAULT NULL,
  `age` int(11) DEFAULT NULL,
  `info` json DEFAULT NULL,
  `tags` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of user_models
-- ----------------------------
BEGIN;
INSERT INTO `user_models` (`id`, `no`, `name`, `email`, `status`, `age`, `info`, `tags`, `created_at`, `updated_at`, `deleted_at`) VALUES (1, '1', 'qwe', 'qwe', 1, 12, '{\"Age\": 18, \"Links\": [\"https://x.com\", \"https://twitch.com\"], \"Address\": \"NY\", \"Verified\": true}', 'tag1,tag2', NULL, NULL, NULL);
INSERT INTO `user_models` (`id`, `no`, `name`, `email`, `status`, `age`, `info`, `tags`, `created_at`, `updated_at`, `deleted_at`) VALUES (2, '2', 'asd', 'asd', 2, 12, '{\"Age\": 18, \"Links\": [\"https://x.com\", \"https://twitch.com\"], \"Address\": \"NY\", \"Verified\": true}', 'tag1,tag2', NULL, NULL, NULL);
INSERT INTO `user_models` (`id`, `no`, `name`, `email`, `status`, `age`, `info`, `tags`, `created_at`, `updated_at`, `deleted_at`) VALUES (3, '3', 'zxc', 'zxc', 1, 12, '{\"Age\": 18, \"Links\": [\"https://x.com\", \"https://twitch.com\"], \"Address\": \"NY\", \"Verified\": true}', 'tag1,tag2', NULL, NULL, NULL);
INSERT INTO `user_models` (`id`, `no`, `name`, `email`, `status`, `age`, `info`, `tags`, `created_at`, `updated_at`, `deleted_at`) VALUES (4, '4', '123', '123', 2, 12, '{\"Age\": 18, \"Links\": [\"https://x.com\", \"https://twitch.com\"], \"Address\": \"NY\", \"Verified\": true}', 'tag1,tag2', NULL, NULL, NULL);
INSERT INTO `user_models` (`id`, `no`, `name`, `email`, `status`, `age`, `info`, `tags`, `created_at`, `updated_at`, `deleted_at`) VALUES (5, '5', 'qaz', 'qaz', 1, 12, '{\"Age\": 18, \"Links\": [\"https://x.com\", \"https://twitch.com\"], \"Address\": \"NY\", \"Verified\": true}', 'tag1,tag2', NULL, NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for videos
-- ----------------------------
DROP TABLE IF EXISTS `videos`;
CREATE TABLE `videos` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `size` int(11) NOT NULL,
  `path` varchar(255) DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Records of videos
-- ----------------------------
BEGIN;
INSERT INTO `videos` (`id`, `user_id`, `size`, `path`, `title`, `created_at`, `updated_at`, `deleted_at`) VALUES (1, 1, 1024, '1.mp4', 'a1', NULL, NULL, NULL);
INSERT INTO `videos` (`id`, `user_id`, `size`, `path`, `title`, `created_at`, `updated_at`, `deleted_at`) VALUES (2, 2, 1024, '2.mp4', 'a2', NULL, NULL, NULL);
INSERT INTO `videos` (`id`, `user_id`, `size`, `path`, `title`, `created_at`, `updated_at`, `deleted_at`) VALUES (3, 3, 1024, '3.mp4', 'a3', NULL, NULL, NULL);
INSERT INTO `videos` (`id`, `user_id`, `size`, `path`, `title`, `created_at`, `updated_at`, `deleted_at`) VALUES (4, 1, 1024, '4.mp4', 'a4', NULL, NULL, NULL);
INSERT INTO `videos` (`id`, `user_id`, `size`, `path`, `title`, `created_at`, `updated_at`, `deleted_at`) VALUES (5, 3, 1024, '5.mp4', 'a5', NULL, NULL, NULL);
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
