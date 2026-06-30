async function test() {
  const q = `
    query {
      userRates(userId: 1, targetType: Anime, limit: 1) {
        id
        status
        score
        anime {
          id
          name
          russian
          genres { name russian }
        }
      }
    }
  `;
  const res = await fetch('https://shikimori.io/api/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'User-Agent': 'App4Every' },
    body: JSON.stringify({ query: q })
  });
  console.log(await res.text());
}
test();
