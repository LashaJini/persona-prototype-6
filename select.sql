WITH days_since_last_scanned AS (
  SELECT
    p.persona_id,
    (EXTRACT(epoch FROM CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::BIGINT - p.last_scanned_at) / 86400 as days_since
  FROM
    persona p
)
SELECT
  p.persona_id,
  p.username,
  --p.social_net_id,
  TO_TIMESTAMP(p.last_scanned_at) AS last,
  TO_TIMESTAMP(p.first_scanned_at) AS first,
  --p.latest_activity_at,
  --p.priority_score,
  --p.total_content_count,
  --p.average_spam,
  --p.average_emotions,
  --p.average_personalities,
  --p.thirdparty,

  --y.yearly_activity_id,
  --y.yearly_activity,

  --w.weekly_activity_id,
  w.sun_to_mon_probs,
  w.sun_to_mon_counts,

  COUNT(c.content_id),
  p.average_spam,

  p.priority_score as p,
  p.priority_score
    + 100 * w.sun_to_mon_probs[1]
    + 
    CASE
      WHEN p.last_scanned_at = 0 THEN 10
      WHEN days_since = 0 THEN -100
      ELSE days_since
    END
    AS redis_score
FROM persona p
INNER JOIN content c USING(persona_id)
INNER JOIN weekly_activity w USING(persona_id)
INNER JOIN yearly_activity y USING(persona_id)
INNER JOIN days_since_last_scanned d USING(persona_id)
GROUP BY
  p.persona_id,
  p.username,
  p.priority_score,
  p.total_content_count,
  w.sun_to_mon_probs,
  w.sun_to_mon_counts,
  d.days_since
ORDER BY
  redis_score DESC;
