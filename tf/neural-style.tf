variable "aws_key_name" {}
variable "access_key" {}
variable "secret_key" {}
 
 
provider "aws" {
    access_key = "${var.access_key}"
    secret_key = "${var.secret_key}"
    region = "eu-west-1"
}
 
resource "aws_spot_instance_request" "neural_style_worker" {
    ami = "ami-640f9417"
    instance_type = "g2.2xlarge"
    availability_zone = "eu-west-1a"
    spot_price = "0.65"
    count = "${var.count}"
    wait_for_fulfillment = "true"
    key_name = "${var.aws_key_name}"
    tags {
        Name = "NeuralStyleWorker"
    }
}
