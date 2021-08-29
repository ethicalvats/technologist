<template>
    <div>
        <h2>Add new todo</h2>
        <form id="add-new-form" v-on:submit="addNewTodo">
            <fieldset>
                <label for="add-new-input">Add new Item</label>
                <input type="text" id="add-new-input" v-model="todoText" />
            </fieldset>
            <input id="submit" type="submit" />
        </form>
    </div>
</template>

<script lang="ts">
import Vue from 'vue'
import service, {Todo} from "./service"
import { v4 as uuidv4 } from "uuid"
export default Vue.extend({
    data(){
        return {
            todoText: ""
        }
    },
    methods:{
        addNewTodo(e: MouseEvent) {
            e.preventDefault()
            const newTodoItem = {
                text: this.todoText.trim(),
                id: uuidv4()
            }
            service.putTodo(newTodoItem)
            .then((todos: Todo[]) => {
                this.$root.$emit('todoItems', todos)
                this.todoText = ""
            })
        }
    }
})
</script>

<style scoped>
    #add-new-form{
        max-width: 200px;
    }
    #add-new-input {
        margin-top: 10px;
    }
    #submit {
        margin: 20px 0 
    }
</style>