<template>
  <!-- RouterView рендерит нужный компонент в зависимости от URL -->
  <!-- transition добавляет плавную анимацию при смене страниц -->
  <RouterView v-slot="{ Component }">
    <Transition name="page" mode="out-in">
      <component :is="Component" />
    </Transition>
  </RouterView>
</template>
