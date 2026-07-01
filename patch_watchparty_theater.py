import re

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'r') as f:
    content = f.read()

# 1. Update height for player wrapper in theater mode
content = content.replace(
'''.room-content.theater-mode .player-wrapper {
  flex: none;
  height: 80vh;
  min-height: unset;
}''',
'''.room-content.theater-mode .player-wrapper {
  flex: none;
  height: 75vh;
  max-height: calc(100vh - 120px);
  min-height: unset;
}''')

# 2. Hide sidebar-media-info in theater mode
content = content.replace(
'''.room-content.theater-mode .sidebar-media-info {
  flex: 1;
  min-width: 300px;
  flex-direction: row;
  align-items: flex-start;
  gap: 16px;
}''',
'''.room-content.theater-mode .sidebar-media-info {
  display: none !important;
}''')

# 3. Insert a copy of media-info into main-area for theater mode
media_info_html = '''
        <!-- Инфо для Theater Mode -->
        <div class="theater-media-info glass animate-fade-in mt-4 p-4 flex gap-4" v-if="isTheaterMode && hasAnimeMetadata && !isLoadingShikimoriDetails">
          <img v-if="shikimoriDetails?.image?.original" :src="'https://shikimori.one' + shikimoriDetails.image.original" class="w-[120px] rounded-md shadow-lg flex-shrink-0" style="align-self: flex-start;">
          <div v-else class="w-[120px] h-[160px] rounded-md bg-white/5 flex items-center justify-center text-2xl flex-shrink-0">📺</div>
          <div class="flex-1 min-w-0">
            <h2 class="text-xl font-bold mb-1">{{ metadataTitle }}</h2>
            <div class="text-sm text-gray-400 mb-3">{{ metadataTitleEn }}</div>
            <div class="flex flex-wrap gap-1 mb-3">
              <span class="badge bg-violet-500/20 text-violet-300 text-xs px-2 py-0.5 rounded">{{ shikimoriDetails?.kind?.toUpperCase() }}</span>
              <span class="badge bg-green-500/20 text-green-300 text-xs px-2 py-0.5 rounded">★ {{ shikimoriDetails?.score }}</span>
              <span v-if="shikimoriDetails?.episodes" class="badge bg-gray-500/20 text-gray-300 text-xs px-2 py-0.5 rounded">{{ shikimoriDetails?.episodes }} эп.</span>
            </div>
          </div>
        </div>
'''

content = content.replace(
    '<!-- Информация о медиа (Только описание и персонажи) -->',
    media_info_html + '\n        <!-- Информация о медиа (Только описание и персонажи) -->'
)

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'w') as f:
    f.write(content)

