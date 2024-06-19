const { expect } = require("chai");

describe("Messaging Contract", function () {
    let Messaging;
    let messaging;
    let owner;
    let addr1;
    let addr2;
    let addrs;

    beforeEach(async function () {
        Messaging = await ethers.getContractFactory("Messaging");
        [owner, addr1, addr2, ...addrs] = await ethers.getSigners();
        messaging = await Messaging.deploy();
        await messaging.deployed();
    });

    it("Should send a message and emit an event", async function () {
        const sendTx = await messaging.connect(addr1).sendMessage(addr2.address, "Hello, Addr2!");

        // Wait for the transaction to be mined
        await sendTx.wait();

        // Check if the event was emitted
        await expect(sendTx)
            .to.emit(messaging, 'MessageSent')
            .withArgs(addr1.address, addr2.address, "Hello, Addr2!");

        // Check if the message is stored in the inbox and dialogues
        const inbox = await messaging.getInbox(addr2.address);
        expect(inbox.length).to.equal(1);
        expect(inbox[0].sender).to.equal(addr1.address);
        expect(inbox[0].receiver).to.equal(addr2.address);
        expect(inbox[0].content).to.equal("Hello, Addr2!");

        const dialogue = await messaging.getDialogue(addr1.address, addr2.address);
        expect(dialogue.length).to.equal(1);
        expect(dialogue[0].sender).to.equal(addr1.address);
        expect(dialogue[0].receiver).to.equal(addr2.address);
        expect(dialogue[0].content).to.equal("Hello, Addr2!");
    });

    it("Should not allow sending a message to oneself", async function () {
        await expect(
            messaging.connect(addr1).sendMessage(addr1.address, "Hello, myself!")
        ).to.be.revertedWith("Cannot send message to yourself");
    });

    it("Should measure gas usage for sendMessage and display results", async function () {
        const gasUsage = [];

        const tx1 = await messaging.connect(addr1).sendMessage(addr2.address, "Hello, Addr2!");
        const receipt1 = await tx1.wait();
        gasUsage.push({ description: "Send message from addr1 to addr2", gasUsed: receipt1.gasUsed.toString() });

        const tx2 = await messaging.connect(addr2).sendMessage(addr1.address, "Hi, Addr1!");
        const receipt2 = await tx2.wait();
        gasUsage.push({ description: "Send message from addr2 to addr1", gasUsed: receipt2.gasUsed.toString() });

        // Output the gas usage table
        console.log("\nGas Usage:");
        console.table(gasUsage);
    });
});
