-- Normalizuje istniejące slugi poradników do docelowego formatu bez sufiksu _v1.
-- Robi to tylko wtedy, gdy bazowy slug nie istnieje jeszcze jako osobny rekord.

UPDATE guides g
SET slug = regexp_replace(g.slug, '_v1$', '')
WHERE g.slug ~ '_v1$'
  AND NOT EXISTS (
    SELECT 1
    FROM guides existing
    WHERE existing.slug = regexp_replace(g.slug, '_v1$', '')
  );