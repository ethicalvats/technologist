import { BaseController } from './BaseController'
import { injectable, inject } from 'inversify';
import * as express from 'express'
import { interfaces, controller, httpPost } from 'inversify-express-utils';
import { TodoService } from '../service/TodoService';

import TYPES from '../types/types'

/**
 * Controller class for Todo API
 * 
 * @export
 * @class TodoController
 * @extends {BaseController}
 * @implements {interfaces.Controller}
 */


@controller('/todos')
export class TodoController extends BaseController implements interfaces.Controller {

    private todoservice: TodoService

    /**
     * Creates an instance of SchemaTestController.
     * @constructor SchemaTestController
     */
    constructor( @inject(TYPES.TodoService) todoservice: TodoService) {
        super()
        this.todoservice = todoservice

    }


    /**
     * Create a new Todo with passed req body params defined as in the json schema
     * 
     * @param {express.Request} req 
     * @param {express.Response} res 
     * @param {express.NextFunction} next 
     * @returns 
     * @memberof TodoController
     */
    @httpPost('/')
    public async createTodo(req: express.Request, res: express.Response, next: express.NextFunction) {

        const body = req.body
        let createdTodo
        try {
            createdTodo = await this.todoservice.createTodo(body)
        } catch (error) {
            return this.renderError(req, res, error)
        }
        this.renderJSON(req, res, { todos: createdTodo }, 201)
    }

}