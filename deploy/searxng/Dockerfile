# SearXNG for sonar.facile.studio, with its settings baked in so a redeploy
# always runs the file in this directory. The settings live outside
# /etc/searxng: that path is a declared volume, and compose keeps an anonymous
# volume across recreations, which would pin the first settings forever.
# Tags: https://hub.docker.com/r/searxng/searxng/tags
FROM docker.io/searxng/searxng:2026.11.14-a7128d5d1
COPY --chown=977:977 settings.yml /usr/local/searxng/sonar/settings.yml
ENV __SEARXNG_CONFIG_PATH=/usr/local/searxng/sonar \
    SEARXNG_SETTINGS_PATH=/usr/local/searxng/sonar/settings.yml \
    FORCE_OWNERSHIP=false \
    GRANIAN_LOG_ACCESS_ENABLED=false
USER 977:977
