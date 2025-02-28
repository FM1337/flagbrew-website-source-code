/// <reference types="Cypress" />
describe("GPSS Tests", () => {
  it("Loads the GPSS Page", () => {
    cy.visit("http://localhost:8080/GPSS");

    cy.contains("GPSS");
  });

  it("Can Query For Pokemon", () => {
    cy.contains("Query")
      .parent()
      .find("input")
      .click()
      .type("Jirachi")
      .wait(1000)
      .then(() => {
        console.log(
          cy
            .root()
            .find("body")
            .find(".v-data-iterator")
            .find(".row")
            .children()
            .should("have.length.gt", 0)
        );
      });
  });

  it("Can Download Pokemon", () => {
    let currentDownloads = 0;
    cy.get("div.v-card:eq(1) > .v-card__actions > .caption:first")
      .invoke("text")
      .then((text) => {
        currentDownloads = Number.parseInt(text.trim());
        console.log(currentDownloads);
      })
      .then(() => {
        cy.contains("Download")
          .click()
          .parentsUntil(".v-card")
          .get(".v-card__actions > .caption:first")
          .invoke("text")
          .then((text) => {
            expect(Number.parseInt(text.trim())).to.equal(currentDownloads + 1);
          });
      });
  });
});
