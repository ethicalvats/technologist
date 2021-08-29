<template>
  <div id="app">
    <list-todo />
    <add-todo />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import AddTodo from './components/AddTodo.vue';
import ListTodo from './components/ListTodo.vue';

@Component({
  components: {
    AddTodo,
    ListTodo
  },
})
export default class App extends Vue {}
</script>
