import axios from "axios"

export type Todo = {
    text: string,
    id: string
}

const putTodo = (item: Todo) : Promise<Todo[]> => {
    let todos: Todo[] = []

    return new Promise((resolve, reject) => {
        axios.post("http://localhost:8080/api/v1/todos", item)
        .then(res => {
            const data = res.data
            if(data){
                todos = data.todos
            }
        })
        .catch(error => {
            reject(error)
        })
        .finally(() => {
            resolve(todos)
        })
    })
}

export default {
    putTodo
}