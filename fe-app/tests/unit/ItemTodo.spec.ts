import { mount, shallowMount } from '@vue/test-utils'
import ItemTodo from "@/components/ItemTodo.vue"


describe('ItemTodo.vue', () => {

    const newTodoItem = "buy some milk"

    it('renders todo item', () => {
      const wrapper = shallowMount(ItemTodo, {
        propsData:{
          item: {
            text: newTodoItem
          }
        }
      })

      const listItem = wrapper.find(".list-item")
      
      expect(listItem.text()).toBe(newTodoItem)
    })
  })