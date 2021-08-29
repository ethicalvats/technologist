import axios from 'axios'
import { injectable, inject } from 'inversify'

import {Todo} from "../model/Todo"

/**
 * Interface for services on Todo API
 * 
 * @export
 * @interface TodoService
 */
export interface TodoService {
    createTodo(demoData: Todo): Promise<Todo[] | Error>
}

/**
 * TodoService interface implementation
 * 
 * @export
 * @class TodoServiceImpl
 * @implements {TodoService}
 */
@injectable()
export class TodoServiceImpl implements TodoService {

    /**
     * Creates a new Todo 
     * 
     * @param {Todo} demoData 
     * @returns {(Promise<Todo|Error>)} 
     * @memberof TodoServiceImpl
     */
    public async  createTodo(todoData: any): Promise<Todo[] | Error> {

        let todo = new Todo(todoData.text, todoData.id)
        const data = {
            "text": todo.getText(),
            "id": todo.getId()
        }

        return new Promise((resolve, reject) => {
            axios.post("http://localhost:9000/insert", data)
            .then(res => {
                const data = res.data
                let todos: Todo[] = []
                data.split("\n").forEach(d => {
                    let buf = Buffer.from(d, 'base64')
                    let bufToString = buf.toString()
                    if(bufToString.length !== 0) {
                        let todo = JSON.parse(buf.toString())
                        todos.push(new Todo(todo.text, todo.id))
                    }
                })
                resolve(todos)
            })
            .catch(error => {
                reject(error)
            })
        })
    }
}