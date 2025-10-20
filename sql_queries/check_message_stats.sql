SELECT 
    id,
    first_name,
    username,
    message_count,
    created_at,
    last_seen
FROM users
ORDER BY message_count DESC;

SELECT 
    id,
    first_name,COALESCE(username, 'N/A') as username,
    message_count as "Total Messages"
FROM users
WHERE message_count > 0
ORDER BY message_count DESC
LIMIT 5;

SELECT 
    COUNT(*) as "Total Users",
    SUM(message_count) as "Total Messages",
    AVG(message_count) as "Avg Messages per User", MAX(message_count) as "Max Messages"
FROM users;

SELECT 
    COUNT(*) as "Users with 0 messages"
FROM users
WHERE message_count = 0;

SELECT  DATE(created_at) as "Date", COUNT(*) as "New Users", SUM(message_count) as "Messages from users registered this day"
FROM users
GROUP BY DATE(created_at)
ORDER BY DATE(created_at) DESC
LIMIT 10;
