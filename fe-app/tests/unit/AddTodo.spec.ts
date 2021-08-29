import { shallowMount, Wrapper, mount } from '@vue/test-utils'
import Vue from 'vue';
import AddTodo from "@/components/AddTodo.vue"
import service, {Todo} from '@/components/service';

describe('AddTodo.vue', () => {

    var wrapper: Wrapper<Vue, Element>;
  
    beforeEach(() => {
      wrapper = mount(AddTodo, {})
    })
  
    it('renders input field', () => {
      const input = wrapper.find("#add-new-input")
      expect(input.element.id).toBe("add-new-input")
    })
  
    it('renders form with submit button', () => {
      const form = wrapper.find("#add-new-form")
      const formSubmit = form.find("input[type=submit]")
      expect(formSubmit.attributes().type).toBe("submit")
    })

    it('calls the backend service to send request', () => {
      const putTodoSpy = jest.spyOn(service, 'putTodo').mockImplementation(() => {
        console.log('mocked putTodo called')
        const todos: Todo[] = []
        return new Promise(resolve =>{
          resolve(todos)
        })
      })
      const form = wrapper.find("#add-new-form")
      form.trigger('submit')
      expect(putTodoSpy).toHaveBeenCalled()
    })
  
  })