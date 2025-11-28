Determine a plan for implementing the specified bead. If a bead was not specified, show a list of open beads with the label `requires-human-review` in s numbered list. Then ask the user to type the number (or bead id) of the one to review.

Your goal is to read the bead, and any related beads if you need more context, then research the code and create a plan to implement the bead and present it to the user for review.
If you need further clarification on anything, ask the user those questions so you have the full context to produce an optimal plan.
The user will likely have suggestions or questions. So iterate with them to help them arrive at the best solution.

Remember to design any user experience changes to be "Delightful" for the user. 

When designing code, favor simplicity, intuitiveness, and functionality. Use good design principles to reduce duplication, keep code clean, and keep it simple. 

Ensure you use the libraries and frameworks that are available to you. Use context7 mcp to understand their capabilities so you are efficient and leveraging those libraries if needed.
Ensure you include full testing of the changed code or any new code.
Ensure you update documentation if applicable.

Once the user approves the plan:
1. Write the detailed plan to the bead
2. Label it `human-reviewed` 
3. Remove the label `requires-human-review`

