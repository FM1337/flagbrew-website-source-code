<template>
  <v-container>
    <h2>Privacy Policy</h2>
    <p>
      Your privacy is important to us. It is FlagBrew&#39;s policy to respect your privacy and comply with any applicable law and regulation
      regarding any personal information we may collect about you, including across our website,
      <a href="https://flagbrew.org">https://flagbrew.org</a>, and other sites we own and operate.
    </p>
    <p>This policy is effective as of 28 March 2021 and was last updated on 28 March 2021.</p>
    <h3>Information We Collect</h3>
    <p>
      Information we collect includes both information you knowingly and actively provide us when using or participating in any of our
      services and promotions, and any information automatically sent by your devices in the course of accessing our products and services.
    </p>
    <h4>Log Data</h4>
    <p>
      When you visit our website, our servers may automatically log the standard data provided by your web browser. It may include your
      device’s Internet Protocol (IP) address, your browser type and version, the pages you visit, the time and date of your visit, the time
      spent on each page, other details about your visit, and technical details that occur in conjunction with any errors you may encounter.
    </p>
    <p>
      Please be aware that while this information may not be personally identifying by itself, it may be possible to combine it with other
      data to personally identify individual persons.
    </p>
    <h4>Collection and Use of Information</h4>
    <p>We may collect personal information from you when you do any of the following on our website:</p>
    <ul>
      <li>Use a mobile device, web browser or homebrew application to access our content</li>
    </ul>
    <p>
      We may collect, hold, use, and disclose information for the following purposes, and personal information will not be further processed
      in a manner that is incompatible with these purposes:
    </p>
    <ul>
      <li>Moderation of GPSS</li>
    </ul>
    <h4>Security of Your Personal Information</h4>
    <p>
      When we collect and process personal information, and while we retain this information, we will protect it within commercially
      acceptable means to prevent loss and theft, as well as unauthorized access, disclosure, copying, use, or modification.
    </p>
    <p>
      Although we will do our best to protect the personal information you provide to us, we advise that no method of electronic
      transmission or storage is 100% secure, and no one can guarantee absolute data security. We will comply with laws applicable to us in
      respect of any data breach.
    </p>
    <h4>How Long We Keep Your Personal Information</h4>
    <p>
      We keep your personal information only for as long as we need to. This time period may depend on what we are using your information
      for, in accordance with this privacy policy. If your personal information is no longer required, we will delete it or make it
      anonymous by removing all details that identify you.
    </p>
    <p>
      However, if necessary, we may retain your personal information for our compliance with a legal, accounting, or reporting obligation or
      for archiving purposes in the public interest, scientific, or historical research purposes or statistical purposes.
    </p>
    <h3>Children’s Privacy</h3>
    <p>
      We do not aim any of our products or services directly at children under the age of 13, and we do not knowingly collect personal
      information about children under 13.
    </p>
    <h3>International Transfers of Personal Information</h3>
    <p>
      The personal information we collect is stored and/or processed where we or our partners, affiliates, and third-party providers
      maintain facilities. Please be aware that the locations to which we store, process, or transfer your personal information may not have
      the same data protection laws as the country in which you initially provided the information. If we transfer your personal information
      to third parties in other countries: (i) we will perform those transfers in accordance with the requirements of applicable law; and
      (ii) we will protect the transferred personal information in accordance with this privacy policy.
    </p>
    <h3>Your Rights and Controlling Your Personal Information</h3>
    <p>
      You always retain the right to withhold personal information from us, with the understanding that your experience of our website may
      be affected. We will not discriminate against you for exercising any of your rights over your personal information. If you do provide
      us with personal information you understand that we will collect, hold, use and disclose it in accordance with this privacy policy.
      You retain the right to request details of any personal information we hold about you.
    </p>
    <p>
      If you believe that any information we hold about you is inaccurate, out of date, incomplete, irrelevant, or misleading, please
      contact us using the details provided in this privacy policy. We will take reasonable steps to correct any information found to be
      inaccurate, incomplete, misleading, or out of date.
    </p>
    <p>
      If you believe that we have breached a relevant data protection law and wish to make a complaint, please contact us using the details
      below and provide us with full details of the alleged breach. We will promptly investigate your complaint and respond to you, in
      writing, setting out the outcome of our investigation and the steps we will take to deal with your complaint. You also have the right
      to contact a regulatory body or data protection authority in relation to your complaint.
    </p>
    <h3>Use of Cookies</h3>
    <p>
      We use &ldquo;cookies&rdquo; to collect information about you and your activity across our site. A cookie is a small piece of data
      that our website stores on your computer, and accesses each time you visit, so we can understand how you use our site. This helps us
      serve you content based on preferences you have specified.
    </p>
    <h3>Limits of Our Policy</h3>
    <p>
      Our website may link to external sites that are not operated by us. Please be aware that we have no control over the content and
      policies of those sites, and cannot accept responsibility or liability for their respective privacy practices.
    </p>
    <h3>Changes to This Policy</h3>
    <p>
      At our discretion, we may change our privacy policy to reflect updates to our business processes, current acceptable practices, or
      legislative or regulatory changes. If we decide to change this privacy policy, we will post the changes here at the same link by which
      you are accessing this privacy policy.
    </p>
    <p>
      If required by law, we will get your permission or give you the opportunity to opt in to or opt out of, as applicable, any new uses of
      your personal information.
    </p>
    <h3>Contact Us</h3>
    <p>For any questions or concerns regarding your privacy, you may contact us using the following details:</p>
    <p>
      FMCore<br />
      [REDACTED]
    </p>
    <v-divider />
    <v-btn color="success" @click="toggleOpt">Opt out of/in for Google Analytics</v-btn>
  </v-container>
</template>

<script>
export default {
  name: "privacy",
  methods: {
    toggleOpt() {
      if (localStorage.getItem("ga-opt-out") === "true") {
        localStorage.setItem("ga-opt-out", false);
        alert("You have opted back in for Google Analytics");
      } else {
        localStorage.setItem("ga-opt-out", true);
        alert("You have opted Out of Google Analytics");
      }
      location.reload();
    },
  },
};
</script>

<style></style>
