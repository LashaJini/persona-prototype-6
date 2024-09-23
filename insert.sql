DO $$
  DECLARE pid uuid;
  BEGIN
    DELETE FROM persona CASCADE;

    INSERT INTO persona (
      username,
      social_net_id
    )
    VALUES (
      'BeautifulMandarin123',
      --'DogDrools',
      --'CantStopPoppin',
      --'AutoModerator',
      --'tjxnlvx',
      1
    )
    RETURNING persona_id INTO pid;

    INSERT INTO yearly_activity (
      persona_id
    )
    VALUES (
      pid
    );

    INSERT INTO weekly_activity (
      persona_id
    )
    VALUES (
      pid
    );
END $$;

SELECT
  p.persona_id,
  --p.username,
  --p.social_net_id,
  --p.last_scanned_at,
  p.first_scanned_at,
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
  --w.sun_to_mon_probs,
  --w.sun_to_mon_counts,

  TO_TIMESTAMP(p.last_scanned_at) AS last,
  p.priority_score
    + 100 * w.sun_to_mon_probs[5]
    + CASE WHEN p.last_scanned_at = 0 THEN 10 ELSE (EXTRACT(epoch FROM CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::BIGINT - p.last_scanned_at) / 86400 END
    AS redis_score
FROM persona p
INNER JOIN weekly_activity w USING(persona_id)
INNER JOIN yearly_activity y USING(persona_id)

ORDER BY
  redis_score DESC
LIMIT 1000
