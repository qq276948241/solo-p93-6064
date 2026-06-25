USE appliance_recycle;

ALTER TABLE appointments ADD COLUMN images TEXT DEFAULT NULL COMMENT '家电图片URL，JSON数组' AFTER status;
