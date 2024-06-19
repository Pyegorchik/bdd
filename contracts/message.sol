contract Messaging {
    struct Message {
        address sender;
        address receiver;
        string content;
    }
    mapping(address => Message[]) public inbox;
    mapping(address => mapping(address => Message[])) public dialogues;
    event MessageSent(
        address indexed sender,
        address indexed receiver,
        string content
    );

    function sendMessage(address _receiver, string memory _content) public {
        require(_receiver != msg.sender, "Cannot send message to yourself");
        Message memory newMessage = Message(msg.sender, _receiver, _content);
        inbox[_receiver].push(newMessage);
        dialogues[msg.sender][_receiver].push(newMessage);
        emit MessageSent(msg.sender, _receiver, _content);
    }

    function getInbox(address user) public view returns (Message[] memory) {
        return inbox[user];
    }

    function getDialogue(
        address user1,
        address user2
    ) public view returns (Message[] memory) {
        return dialogues[user1][user2];
    }
}
