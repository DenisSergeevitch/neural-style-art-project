package net.franchu.neuralstyleartproject;

import com.google.common.primitives.Bytes;

import net.franchu.neuralstyleartproject.nano.Imager;

import java.net.ConnectException;
import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

/**
 * Created by mgilbir on 09/03/16.
 */
public class ImagerClient {
    private final ManagedChannel channel;
    private final NeuralStyleImagerGrpc.NeuralStyleImagerBlockingStub blockingStub;
    private final NeuralStyleImagerGrpc.NeuralStyleImagerStub asyncStub;

    /**
     * Construct client for accessing RouteGuide server at {@code host:port}.
     */
    public ImagerClient(String host, int port) {
        this(ManagedChannelBuilder.forAddress(host, port).usePlaintext(true));
    }

    public ImagerClient(ManagedChannelBuilder<?> channelBuilder) {
        channel = channelBuilder.build();
        blockingStub = NeuralStyleImagerGrpc.newBlockingStub(channel);
        asyncStub = NeuralStyleImagerGrpc.newStub(channel);
    }

    public void shutdown() throws InterruptedException {
        channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
    }

    public void createJob(String name, byte[] image) {
        Imager.CreateJobRequest request = new Imager.CreateJobRequest();
        request.name = name;
        net.franchu.neuralstyleartproject.nano.Image.InputImage content = new net.franchu.neuralstyleartproject.nano.Image.InputImage();

        int format = net.franchu.neuralstyleartproject.nano.Image.JPG;

        content.format = format;
        content.image = image;

        request.content = content;

        Imager.CreateJobResponse response = blockingStub.createJob(request);
    }
}
