# RabbitMQ and message queueing

*Message-oriented middleware*

Используя *RabbitMQ* (или другую аналогичную по функциональности систему обмена сообщениям) продемонстрировать асинхронное взаимодействие нескольких систем/подсистем (приложений) на основе очереди.

**Материалы по RabbitMQ**:

   * https://www.rabbitmq.com/getstarted.html
   * https://www.cloudamqp.com/blog/2015-05-18-part1-rabbitmq-for-beginners-what-is-rabbitmq.html

*Starting server:* `docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management`

1. Реализовать два варианта очереди: **Producer/Consumer** (Point-to-Point) и **Publish/Subscribe** (Topic).
   
   - **ОТВЕТ:** *примеры показаны в консольном приложении, по командам pc-send, pc-receive, ps-send, ps-receive.*

2. Реализовать следующую логику - клиент отправляет сообщение в очередь, один из консьюмеров его вычитывает, модифицирует и кладет в ответную очередь клиенту, который выполнял отправку, клиент вычитывает ответ и отображает его.

   - **ОТВЕТ:** *пример показан в консольном приложении, по команде* **start**. *По адресу 127.0.0.1:8080/ доступны методы GET, POST для /message, /topic (Producer/Consumer и Publish/Subscribe соответственно). Метод POST пишет в "PRODUCER/PUBLISHER" (сервис читает его, модифицирует и пишет в "COMSUMER/SUBSCRIBER"); метод GET читает из "COMSUMER/SUBSCRIBER".*

3. Показать и настроить варианты предоставляемые MOM, связанные с:
   * подтверждением доставки/обработки сообщений клиентом (Message Acknowledgment);
   * сохранности очереди сообщений (Message Persistence - Durable queue);
   * время пребывания сообщения в очереди (Message TTL);
   * максимальная длина очереди (Max length) (что происходит с сообщениями когда очередь заполнена).


    - **ОТВЕТЫ:**

    `rabbit.queue.Messages = 2 // count of messages not awaiting acknowledgment`

```
    rabbit.queue, err = rabbit.channel.QueueDeclare(
		rabbit.QueueName, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		amqp.Table{ // arguments
			"x-message-ttl": int32(60000), // Declares a queue with the x-message-ttl extension
			// to exercise integer serialization. 60 sec.)
			"x-max-length": 10, // Maximum number of messages can be set
			// by supplying the `x-max-length` queue declaration argument with a non-negative integer value.
		})
```

4. Для варианта **Producer/Consumer** показать случай, когда *Consumer* берет из очереди сообщение на обработку, но не может его обработать: падает/не возвращает Ack/возвращает негативный Ack. Показать, будет ли при этом данное необработанное сообщение взято на обработку другим *Consumer* или окажется потерянным.

   -  **ОТВЕТ:** в процессе..
