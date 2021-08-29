<template>
    <div>
        <h1>Todo List</h1>
        <ul id="todo-list">
            <item-todo v-for="item in items" :key="item.id" :item="item" />
        </ul>
    </div>
</template>

<script lang="ts">
import Vue from 'vue'
import ItemTodo from "./ItemTodo.vue"
import {Todo} from "./service"

type ListTodoProps = {
    items: Todo[]
}

export default Vue.extend({
    components:{
        ItemTodo
    },
    data() {
        let propsData: ListTodoProps = {
            items: []
        }
        return propsData
        
    },
    mounted(){
        this.$root.$on('todoItems', (todos: Todo[]) => {
            this.update(todos)
        })
    },
    methods: {
        update(todos: Todo[]){
            this.items = todos
        }
    }
})
</script>