USE albiondata;

/** TABLE market_stats **/
DROP TABLE IF EXISTS `market_stats`;

CREATE TABLE `market_stats` (
  `id` int(10) UNSIGNED NOT NULL,
  `item_id` varchar(255) NOT NULL,
  `location` int(11) NOT NULL,
  `price_min` int(11) DEFAULT NULL,
  `price_max` int(11) DEFAULT NULL,
  `price_avg` int(11) DEFAULT NULL,
  `timestamp` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

ALTER TABLE `market_stats`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `market_stats`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=195;

ALTER TABLE `market_stats` ADD UNIQUE `item_id_location_timestamp_unique` (`item_id`, `location`, `timestamp`) USING BTREE;


DROP PROCEDURE IF EXISTS `execute_stmt`;

DELIMITER $$
CREATE PROCEDURE `execute_stmt`(IN sql_text TEXT)
    MODIFIES SQL DATA

    BEGIN
        DECLARE EXIT HANDLER FOR SQLEXCEPTION
        BEGIN
                SELECT CONCAT(sql_text, ' is not valid');
        END;
        SET @SQL := sql_text;
        PREPARE stmt FROM @SQL;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END;
$$

DELIMITER ;



/** PROCEDURE `create_hour_stats` creates stats in all locations for the given hour
2017-08-28T19:00:00Z00:00
 **/
DROP PROCEDURE IF EXISTS `create_hour_stats`;

DELIMITER $$
CREATE PROCEDURE `create_hour_stats`(IN var_timestamp timestamp)
    MODIFIES SQL DATA

    BLOCK1: BEGIN
        DECLARE var_hour_from VARCHAR(20);
        DECLARE var_hour_to VARCHAR(20);
        SELECT DATE_FORMAT(var_timestamp, '%Y-%m-%dT%H:00:00Z00:00') INTO var_hour_from;
        SELECT DATE_FORMAT(var_timestamp, '%Y-%m-%dT%H:59:59Z00:00') INTO var_hour_to;

        BLOCK2: BEGIN
            DECLARE var_location int;
            DECLARE var_item_id VARCHAR(255);
            DECLARE var_price_min int;
            DECLARE var_price_max int;
            DECLARE var_price_avg decimal;
            DECLARE cursor_items_done boolean;

            DECLARE cursor_items CURSOR FOR SELECT `location`, `item_id` FROM `market_orders` WHERE `auction_type` = 'offer' AND (`updated_at` >= var_hour_from AND `updated_at` <= var_hour_to) GROUP BY `location`, `item_id`;
            DECLARE CONTINUE HANDLER FOR NOT FOUND SET cursor_items_done = 1;

            START TRANSACTION;
            OPEN cursor_items;
            Reading_Items_Loop: LOOP
                FETCH cursor_items INTO var_location, var_item_id;
                IF cursor_items_done THEN
                    CLOSE cursor_items;
                    LEAVE Reading_Items_Loop;
                END IF;  

                BLOCK3: BEGIN
                    DECLARE CONTINUE HANDLER FOR NOT FOUND SET var_price_min = 0;
                    SELECT `price` FROM `market_orders` WHERE `location` = var_location AND `item_id` = var_item_id AND `auction_type` = 'offer' AND  (`updated_at` >= var_hour_from AND `updated_at` <= var_hour_to) ORDER BY `updated_at`, `price` LIMIT 1 INTO var_price_min;
                END BLOCK3;

                BLOCK3: BEGIN
                    DECLARE CONTINUE HANDLER FOR NOT FOUND SET var_price_max = 0;
                    SELECT `price` FROM `market_orders` WHERE `location` = var_location AND `item_id` = var_item_id AND `auction_type` = 'offer' AND (`updated_at` >= var_hour_from AND `updated_at` <= var_hour_to) ORDER BY `updated_at` ASC, `price` DESC LIMIT 1 INTO var_price_max;                    
                END BLOCK3;

                BLOCK3: BEGIN
                    DECLARE CONTINUE HANDLER FOR NOT FOUND SET var_price_avg = 0.0;
                    SELECT AVG(`price`) as 'price_avg' FROM `market_orders` WHERE `location` = var_location AND `item_id` = var_item_id AND `auction_type` = 'offer' AND (`updated_at` >= var_hour_from AND `updated_at` <= var_hour_to) INTO var_price_avg;                    
                END BLOCK3;

                Set @stmt = CONCAT('INSERT IGNORE INTO `market_stats` (`item_id`, `location`, `timestamp`, `price_min`, `price_max`, `price_avg`) VALUES (\'', IFNULL(var_item_id, ''), '\',', IFNULL(var_location, 0), ',\'', IFNULL(var_hour_from, 0), '\',', IFNULL(var_price_min, 0), ',', IFNULL(var_price_max, 0), ',', IFNULL(var_price_avg, 0.0), ');');
                CALL `execute_stmt`(@stmt);
            END LOOP;

            COMMIT;
            
        END BLOCK2;
    END BLOCK1;

$$

DELIMITER ;


/** PROCEDURE `create_now_stats` creates for stats for the last hour in all locations **/
DROP PROCEDURE IF EXISTS `create_now_stats`; 

DELIMITER $$
CREATE PROCEDURE `create_now_stats`() 
	MODIFIES SQL DATA
    
    CALL `create_hour_stats`(UNIX_TIMESTAMP());
$$

DELIMITER ;

/** PROCEDURE `create_all_data_stats` A VERY EXPENSIVE script that creates stats for all items in the DB. **/
DROP PROCEDURE IF EXISTS `create_all_data_stats`; 

DELIMITER $$
CREATE PROCEDURE `create_all_data_stats`() 
	MODIFIES SQL DATA
    
BLOCK1: BEGIN
    DECLARE var_timestamp timestamp;
    DECLARE cursor_done boolean;
    DECLARE cursor_all CURSOR FOR SELECT DISTINCT(DATE_FORMAT(`updated_at`, '%Y-%m-%dT%H:00:00Z00:00')) FROM `market_orders`;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET cursor_done = TRUE;

    OPEN cursor_all;
    Reading_Locations: LOOP
        FETCH cursor_all INTO var_timestamp;
        IF cursor_done THEN
            LEAVE Reading_Locations;
        END IF;

        CALL `create_hour_stats`(var_timestamp);
    END LOOP;
    CLOSE cursor_all;    
END BLOCK1;
$$

DELIMITER ; 