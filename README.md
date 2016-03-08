# Art project at Elckerlyc

## Installation

To automate the instance creation and destruction we use [terraform](https://terraform.io). If you are on OSX and use homebrew, you can get it with `brew install terraform`

Once that is done, you can check out this project, go to the terraform directory and create a new file called `terraform.tfvars` with the following content:

	# EC2 Variables
	access_key = "XXXXXX"
	secret_key = "XXXXXX"
	aws_key_name = "XXXXXX"

Where you will replace the XXXXXX with the correct values for your account.

To deploy the instance, run: `terraform apply`

After it completes, with `terraform show` you can see the IP where your brand new instance is listening and you can run the commands in the [neural-style](https://github.com/jcjohnson/neural-style) under the directory `neural-style`

An example of the command that you can run would be:
`th neural_style.lua -style_image input/blabla.jpg -content_image ~/my_image.jpg -print_iter 1 -gpu 0 -backend cudnn -cudnn_autotune`

You can put images in the instance and take them out with scp.

When you are done and want to tear down the instance, run `terraform destroy`

REMEMBER TO DESTROY YOUR INSTANCES WHEN YOU'RE NOT USING THEM!


## Aspirational TODO list
 - Deploy multiple nodes with neural-style
 - Queue to hold jobs
 - Deploy one job per gpu
 - Save results every n iterations
 - Stream the results to a web browser
 - Resulting image can be exported
 - Resulting image can be exported in a PDF with input + style = result
 - Input via Android app

