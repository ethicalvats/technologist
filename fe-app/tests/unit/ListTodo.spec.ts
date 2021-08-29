import { shallowMount, Wrapper } from '@vue/test-utils'
import Vue from 'vue';
import ListTodo from "@/components/ListTodo.vue"

describe('ListTodo.vue', () => {

    var wrapper: Wrapper<Vue, Element>;
  
    beforeEach(() => {
      wrapper = shallowMount(ListTodo, {})
    })
  
    it('renders list view', () => {
      const list = wrapper.find("#todo-list")
      expect(list.element.tagName == 'ul')
    })
  
  })