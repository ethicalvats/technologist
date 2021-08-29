/**
 * Todo model
 * 
 * @export
 * @class Todo
 */
export class Todo {
    
    /**
     * Creates an instance of Todo.
    
        * @param {string} id 
        * @param {string} text
        * @memberof Todo
        */
    constructor(
        private text: string,
        private id: string
    ) {
    }

    /**
     * 
     * 
     * @readonly
     * @type {string}
     * @memberof Todo
     */
    getText(): string {
        return this.text;
    }

    /**
     * 
     * 
     * @readonly
     * @type {string}
     * @memberof Todo
     */
    getId(): string {
        return this.id;
    }
}