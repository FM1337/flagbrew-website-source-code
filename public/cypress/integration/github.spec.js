describe("Can Load GitHub Project", () => {
  it("Loads the PKSM GitHub project page on the website", () => {
    cy.visit("http://localhost:8080/projects/PKSM");

    cy.contains("Multipurpose and portable save manager for generations III to VIII, programmed in C++.");
  });
});
