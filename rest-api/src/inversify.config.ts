import 'reflect-metadata'
import { Container } from 'inversify';
import TYPES from './types/types';

import { TodoService, TodoServiceImpl } from './service/TodoService'

const container = new Container();
container.bind<TodoService>(TYPES.TodoService).to(TodoServiceImpl)

export default container;