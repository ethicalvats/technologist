// https://docs.cypress.io/api/introduction/api.html

describe('My App', () => {

  it('expects to enter and save new todo', () => {

    var newTodoItem = "buy some milk"

    cy.visit('/')

    cy.get("#add-new-input")
    .type(newTodoItem)

    cy.get("#add-new-form")
    .submit()
    
    cy.get("ul#todo-list li")
    .last()
    .contains(newTodoItem)
  })
})
